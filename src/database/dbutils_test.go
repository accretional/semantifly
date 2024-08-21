package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createTestingDatabase() (*pg.DB, error) {
	// Connect to the default "postgres" database
	db := pg.Connect(&pg.Options{
		User:     "gitpod",
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
		User:     "gitpod",
		Addr:     "localhost:5432",
		Database: "testdb",
	})

	return testDB, nil
}

func removeTestingDatabase() error {
	// Connect to the default "postgres" database to drop the test database
	defaultDB := pg.Connect(&pg.Options{
		User:     "gitpod",
		Addr:     "localhost:5432",
		Database: "postgres",
	})
	defer defaultDB.Close()

	// Drop the test database
	_, err := defaultDB.Exec("DROP DATABASE testdb")
	if err != nil {
		return fmt.Errorf("Failed to drop test database: %v", err)
	}

	return nil
}

func TestEstablishConnection(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	conn, err := establishConnection(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	conn.Close(ctx)
	removeTestingDatabase()
}

func TestInsertRow(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	conn, err := establishConnection(ctx)
	assert.NoError(t, err)
	defer conn.Close(ctx)

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
				WordOccurrences:   map[string]int32{"test": 1},
			},
		},
	}

	err = insertRows(ctx, conn, index)
	assert.NoError(t, err)

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	assert.NoError(t, err)

	removeTestingDatabase()
}

func TestQueryRow(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	conn, err := establishConnection(ctx)
	assert.NoError(t, err)
	defer conn.Close(ctx)

	expectedEntry := &pb.IndexListEntry{
		Name:              "Test Entry",
		URI:               "http://example.com",
		DataType:          pb.DataType_TEXT,
		SourceType:        pb.SourceType_WEBPAGE,
		FirstAddedTime:    timestamppb.Now(),
		LastRefreshedTime: timestamppb.Now(),
		Content:           "Test Content",
		WordOccurrences:   map[string]int32{"test": 1},
	}

	index := &pb.Index{
		Entries: []*pb.IndexListEntry{expectedEntry},
	}

	err = insertRows(ctx, conn, index)
	assert.NoError(t, err)

	entry, err := queryRow(ctx, conn, "Test Entry")

	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, expectedEntry.Name, entry.Name)
	assert.Equal(t, expectedEntry.URI, entry.URI)
	assert.Equal(t, expectedEntry.DataType, entry.DataType)
	assert.Equal(t, expectedEntry.SourceType, entry.SourceType)
	assert.Equal(t, expectedEntry.Content, entry.Content)
	assert.Equal(t, expectedEntry.WordOccurrences, entry.WordOccurrences)

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	assert.NoError(t, err)

	removeTestingDatabase()
}
