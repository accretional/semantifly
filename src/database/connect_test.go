package database

import (
	"context"
	"os"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEstablishConnection(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	ctx := context.Background()
	conn, err := establishConnection(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Clean up
	conn.Close(ctx)
}

func TestInsertRow(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

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

	err = insertRow(ctx, conn, index)
	assert.NoError(t, err)

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	assert.NoError(t, err)
}

func TestQueryRow(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

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

	err = insertRow(ctx, conn, index)
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
}
