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

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)

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
				Name:              "Test Entry",
				URI:               "http://example.com",
				DataType:          pb.DataType_TEXT,
				SourceType:        pb.SourceType_WEBPAGE,
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

	index := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name:              "Test Entry",
				URI:               "http://example.com",
				DataType:          pb.DataType_TEXT,
				SourceType:        pb.SourceType_WEBPAGE,
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

	entry, err := getContentMetadata(ctx, conn, "Test Entry")
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}

	if entry == nil {
		t.Fatalf("entry is nil")
	}

	expectedEntry := index.Entries[0]

	type Metadata struct {
		Name       string
		URI        string
		DataType   pb.DataType
		SourceType pb.SourceType
	}

	expectedMetadata := Metadata{
		Name:       expectedEntry.Name,
		URI:        expectedEntry.URI,
		DataType:   expectedEntry.DataType,
		SourceType: expectedEntry.SourceType,
	}

	actualMetadata := Metadata{
		Name:       entry.Name,
		URI:        entry.URI,
		DataType:   entry.DataType,
		SourceType: entry.SourceType,
	}

	if !reflect.DeepEqual(expectedMetadata, actualMetadata) {
		t.Fatalf("Partial entry mismatch.\nExpected: %+v\nGot: %+v", expectedMetadata, actualMetadata)
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
