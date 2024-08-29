package database

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupPostgres() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory: %v", err)
	}

	if filepath.Base(currentDir) != "semantifly" {
		err = os.Chdir("../..")
		if err != nil {
			return fmt.Errorf("Failed to change directory: %v", err)
		}
	}

	cmd := exec.Command("bash", "setup_postgres.sh")

	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to setup PostgreSQL server: %v", err)
	}

	return nil
}

func createTestingDatabase() (*pg.DB, error) {
	// Connect to the default "postgres" database
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Addr:     "localhost:5432",
		Database: "postgres",
	})

	// Drop the database if it exists, then create it
	_, err := db.Exec("DROP DATABASE IF EXISTS testdb")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to drop existing test database: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE testdb")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create test database: %v", err)
	}

	// Close the connection to the "postgres" database
	db.Close()

	// Connect to the newly created database
	testDB := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Addr:     "localhost:5432",
		Database: "testdb",
	})

	return testDB, nil
}

func removeTestingDatabase() error {
	// Connect to the default "postgres" database to drop the test database
	defaultDB := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Addr:     "localhost:5432",
		Database: "postgres",
	})
	defer defaultDB.Close()

	// Terminate all connections to the test database
	_, err := defaultDB.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'testdb'")
	if err != nil {
		return fmt.Errorf("Failed to terminate connections to test database: %v", err)
	}

	// Drop the test database
	_, err = defaultDB.Exec("DROP DATABASE IF EXISTS testdb")
	if err != nil {
		return fmt.Errorf("Failed to drop test database: %v", err)
	}

	return nil
}

func TestPostgres(t *testing.T) {

	err := setupPostgres()

	if err != nil {
		t.Fatalf("setupPostgres failed: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))

	assert.NoError(t, err)
	defer conn.Close(ctx)

	// Test database table initialisation
	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database schema: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}

func TestProtoIndexCreation(t *testing.T) {

	err := setupPostgres()

	if err != nil {
		t.Fatalf("setupPostgres failed: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))

	assert.NoError(t, err)
	defer conn.Close(ctx)

	// Test database table initialisation
	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database schema: %v", err)
	}

	// Test index creation for wordOccurrence
	err = createProtoFieldIndex(ctx, conn, "entry->'wordOccurrence'")
	if err != nil {
		t.Fatalf("Failed to create a Proto field index: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}

func TestInsertRow(t *testing.T) {

	err := setupPostgres()

	if err != nil {
		t.Fatalf("setupPostgres failed: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to establish connection to the database: %v", err)
	}

	// Test database table initialisation
	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database schema: %v", err)
	}

	index := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name: "Test Entry",
				ContentMetadata: &pb.ContentMetadata{
					URI:        "http://example.com",
					DataType:   pb.DataType_TEXT,
					SourceType: pb.SourceType_WEBPAGE,
				},
				FirstAddedTime:    timestamppb.Now(),
				LastRefreshedTime: timestamppb.Now(),
				Content:           "Test Content",
				WordOccurrences:   map[string]int32{"test": 1, "content": 1},
			},
		},
	}

	err = insertRows(ctx, conn, index)
	if err != nil {
		t.Fatalf("failed to insert rows: %v", err)
	}

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	if err != nil {
		t.Fatalf("failed to delete test entry: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}

func TestQueryRow(t *testing.T) {

	err := setupPostgres()

	if err != nil {
		t.Fatalf("setupPostgres failed: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to establish connection to the database: %v", err)
	}

	// Test database table initialisation
	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database schema: %v", err)
	}

	expectedMetadata := &pb.ContentMetadata{
		URI:        "http://example.com",
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_WEBPAGE,
	}

	index := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name:              "Test Entry",
				ContentMetadata:   expectedMetadata,
				FirstAddedTime:    timestamppb.Now(),
				LastRefreshedTime: timestamppb.Now(),
				Content:           "Test Content",
				WordOccurrences:   map[string]int32{"test": 1, "content": 1},
			},
		},
	}

	err = insertRows(ctx, conn, index)
	if err != nil {
		t.Fatalf("failed to insert rows: %v", err)
	}

	fetchedMetadata, err := getContentMetadata(ctx, conn, "Test Entry")
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}

	if fetchedMetadata == nil {
		t.Fatalf("entry is nil")
	}

	if !reflect.DeepEqual(expectedMetadata, fetchedMetadata) {
		t.Fatalf("Partial entry mismatch.\nExpected: %+v\nGot: %+v", expectedMetadata, fetchedMetadata)
	}

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	if err != nil {
		t.Fatalf("failed to delete test entry: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}

}

func TestDeleteRow(t *testing.T) {

	err := setupPostgres()

	if err != nil {
		t.Fatalf("setupPostgres failed: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to establish connection to the database: %v", err)
	}

	// Test database table initialisation
	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database schema: %v", err)
	}

	index := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name: "Test Entry",
				ContentMetadata: &pb.ContentMetadata{
					URI:        "http://example.com",
					DataType:   pb.DataType_TEXT,
					SourceType: pb.SourceType_WEBPAGE,
				},
				FirstAddedTime:    timestamppb.Now(),
				LastRefreshedTime: timestamppb.Now(),
				Content:           "Test Content",
				WordOccurrences:   map[string]int32{"test": 1, "content": 1},
			},
		},
	}

	// Inserting an index
	err = insertRows(ctx, conn, index)
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	// Fetching the row after inserting it
	_, err = getContentMetadata(ctx, conn, "Test Entry")
	if err != nil {
		t.Fatalf("Failed to query row after insertion: %v", err)
	}

	// Test delete row
	err = deleteRow(ctx, conn, "Test Entry")
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Fetching the row after deleting it
	_, err = getContentMetadata(ctx, conn, "Test Entry")
	if err == nil {
		t.Fatalf("Index entry not deleted.")
	} else if !strings.Contains(err.Error(), "no entry found") {
		t.Fatalf("Failed to query row: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}
