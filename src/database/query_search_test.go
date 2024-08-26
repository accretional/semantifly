package database

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSearchIndexForTopMatches(t *testing.T) {

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

	ctx := context.Background()
	conn, err := establishConnection(ctx)
	if err != nil {
		t.Fatalf("Failed to establish connection to database: %v", err)
	}
	defer conn.Close(ctx)

	err = initializeTables(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to initialise the database tables: %v", err)
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
			{
				Name:              "Test JSON",
				URI:               "http://echo.jsontest.com/title/foo/content/bar",
				DataType:          pb.DataType_TEXT,
				SourceType:        pb.SourceType_WEBPAGE,
				FirstAddedTime:    timestamppb.Now(),
				LastRefreshedTime: timestamppb.Now(),
				Content:           `{"title": "foo","content": "bar"}`,
				WordOccurrences:   map[string]int32{"title": 1, "foo": 1, "content": 1, "bar": 1},
			},
		},
	}

	err = insertRows(ctx, conn, index)
	if err != nil {
		t.Fatalf("Failed to insert rows: %v", err)
	}

	queryIndex, err := searchIndexForTopMatches(ctx, conn, "Test contents", 3)
	if err != nil {
		t.Fatalf("Failed to query row: %v", err)
	}

	if len(queryIndex.Entries) < 1 {
		t.Fatalf("Query output is nil")
	}

	queryOutput := queryIndex.Entries[0]
	expectedEntry := index.Entries[0]

	if expectedEntry.Name != queryOutput.Name {
		t.Fatalf("Expected name %s, got %s", expectedEntry.Name, queryOutput.Name)
	}

	if expectedEntry.URI != queryOutput.URI {
		t.Fatalf("Expected URI %s, got %s", expectedEntry.URI, queryOutput.URI)
	}

	if expectedEntry.DataType != queryOutput.DataType {
		t.Fatalf("Expected DataType %v, got %v", expectedEntry.DataType, queryOutput.DataType)
	}

	if expectedEntry.SourceType != queryOutput.SourceType {
		t.Fatalf("Expected SourceType %v, got %v", expectedEntry.SourceType, queryOutput.SourceType)
	}

	if expectedEntry.Content != queryOutput.Content {
		t.Fatalf("Expected Content %s, got %s", expectedEntry.Content, queryOutput.Content)
	}

	if !reflect.DeepEqual(expectedEntry.WordOccurrences, queryOutput.WordOccurrences) {
		t.Fatalf("Expected WordOccurrences %v, got %v", expectedEntry.WordOccurrences, queryOutput.WordOccurrences)
	}

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	if err != nil {
		t.Fatalf("Failed to delete test entry: %v", err)
	}

}
