package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createTestingDatabase() (*sql.DB, error) {
	// Create the database for testing
	db, err := sql.Open("postgres", "postgres://gitpod@localhost:5432/postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE testdb")
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %v", err)
	}

	return db, nil
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

	// Clean up
	conn.Close(ctx)

	// Drop the test database
	_, err = db.Exec("DROP DATABASE testdb")
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}
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
			},
		},
	}

	err = insertRows(ctx, conn, index)
	assert.NoError(t, err)

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	assert.NoError(t, err)

	// Drop the test database
	_, err = db.Exec("DROP DATABASE testdb")
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}
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

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	assert.NoError(t, err)

	// Drop the test database
	_, err = db.Exec("DROP DATABASE testdb")
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}
}
