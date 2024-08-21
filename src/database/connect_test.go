package database

import (
	"context"
	"os"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEstablishConnection(t *testing.T) {
	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://gitpod@localhost:5432/testdb")

	ctx := context.Background()
	conn, err := establishConnection(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Clean up
	conn.Close(ctx)
}

func TestInsertRow(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS index_list").WillReturnResult(pgxmock.NewResult("CREATE TABLE", 1))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO index_list").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

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

	err = insertRow(context.Background(), mock, index)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestQueryRow(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	expectedEntry := &pb.IndexListEntry{
		Name:              "Test Entry",
		URI:               "http://example.com",
		DataType:          pb.DataType_TEXT,
		SourceType:        pb.SourceType_WEBPAGE,
		FirstAddedTime:    timestamppb.Now(),
		LastRefreshedTime: timestamppb.Now(),
		Content:           "Test Content",
	}

	rows := pgxmock.NewRows([]string{"name", "uri", "data_type", "source_type", "first_added_time", "last_refreshed_time", "content"}).
		AddRow(expectedEntry.Name, expectedEntry.URI, expectedEntry.DataType.String(), expectedEntry.SourceType.String(),
			expectedEntry.FirstAddedTime.AsTime(), expectedEntry.LastRefreshedTime.AsTime(), expectedEntry.Content)

	mock.ExpectQuery("SELECT (.+) FROM index_list WHERE name=\\$1").
		WithArgs("Test Entry").
		WillReturnRows(rows)

	entry, err := queryRow(context.Background(), mock, "Test Entry")
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, expectedEntry.Name, entry.Name)
	assert.Equal(t, expectedEntry.URI, entry.URI)
	assert.Equal(t, expectedEntry.DataType, entry.DataType)
	assert.Equal(t, expectedEntry.SourceType, entry.SourceType)
	assert.Equal(t, expectedEntry.Content, entry.Content)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
