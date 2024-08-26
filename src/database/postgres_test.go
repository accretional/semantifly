package database

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupPostgres() error {

	err := os.Chdir("../..")
	if err != nil {
		return fmt.Errorf("Failed to change directory: %v", err)
	}

	cmd := exec.Command("bash", "setup_postgres.sh")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

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
	const maxRetries = 3
	var err error

	if err := setupPostgres(); err != nil {
		t.Fatalf("Failed to setup Postgres server: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	var db *pg.DB
	for i := 0; i < maxRetries; i++ {
		db, err = createTestingDatabase()
		if err == nil || !strings.Contains(err.Error(), "connection reset by peer") {
			break
		}
		t.Logf("Failed to create test database: Connection Reset by peer. Retrying...")
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		t.Fatalf("Failed to create test database after %d attempts: %v", maxRetries, err)
	}
	defer db.Close()

	// Test connection
	ctx := context.Background()
	conn, err := establishConnection(ctx)

	assert.NoError(t, err)
	defer conn.Close(ctx)

	// Test database table initialisation
	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database schema: %v", err)
	}

	// Test row insertion
	err = testInsertRow(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	// Test query
	err = testQueryRow(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to query row: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}

func testInsertRow(ctx context.Context, conn *pgx.Conn) error {

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

	err := insertRows(ctx, conn, index)
	if err != nil {
		return fmt.Errorf("failed to insert rows: %v", err)
	}

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	if err != nil {
		return fmt.Errorf("failed to delete test entry: %v", err)
	}

	return nil
}

func testQueryRow(ctx context.Context, conn *pgx.Conn) error {

	expectedEntry := &pb.IndexListEntry{
		Name:              "Test Entry",
		URI:               "http://example.com",
		DataType:          pb.DataType_TEXT,
		SourceType:        pb.SourceType_WEBPAGE,
		FirstAddedTime:    timestamppb.Now(),
		LastRefreshedTime: timestamppb.Now(),
		Content:           "Test Content",
		WordOccurrences:   map[string]int32{"test": 1, "content": 1},
	}

	index := &pb.Index{
		Entries: []*pb.IndexListEntry{expectedEntry},
	}

	err := insertRows(ctx, conn, index)
	if err != nil {
		return fmt.Errorf("failed to insert rows: %v", err)
	}

	entry, err := queryRow(ctx, conn, "Test Entry")
	if err != nil {
		return fmt.Errorf("failed to query row: %v", err)
	}

	if entry == nil {
		return fmt.Errorf("entry is nil")
	}

	if expectedEntry.Name != entry.Name {
		return fmt.Errorf("expected name %s, got %s", expectedEntry.Name, entry.Name)
	}

	if expectedEntry.URI != entry.URI {
		return fmt.Errorf("expected URI %s, got %s", expectedEntry.URI, entry.URI)
	}

	if expectedEntry.DataType != entry.DataType {
		return fmt.Errorf("expected DataType %v, got %v", expectedEntry.DataType, entry.DataType)
	}

	if expectedEntry.SourceType != entry.SourceType {
		return fmt.Errorf("expected SourceType %v, got %v", expectedEntry.SourceType, entry.SourceType)
	}

	if expectedEntry.Content != entry.Content {
		return fmt.Errorf("expected Content %s, got %s", expectedEntry.Content, entry.Content)
	}

	if !reflect.DeepEqual(expectedEntry.WordOccurrences, entry.WordOccurrences) {
		return fmt.Errorf("expected WordOccurrences %v, got %v", expectedEntry.WordOccurrences, entry.WordOccurrences)
	}

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	if err != nil {
		return fmt.Errorf("failed to delete test entry: %v", err)
	}

	return nil
}
