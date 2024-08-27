package database

import (
	"context"
	"fmt"
	"os"
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

func TestDatabaseSearch(t *testing.T) {
	const maxRetries = 3
	var err error

	// if err := setupPostgres(); err != nil {
	// 	t.Fatalf("Failed to setup Postgres server: %v", err)
	// }

	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	// Set a mock DATABASE_URL for testing
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

	err = testLexicalSearch(ctx, conn)
	if err != nil {
		t.Fatalf("Failed lexical search: %v", err)
	}

	// Cleanup
	if err := removeTestingDatabase(); err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}

func testLexicalSearch(ctx context.Context, conn *pgx.Conn) error {
	index := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name:              "Test Example",
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

	err := insertRows(ctx, conn, index)
	if err != nil {
		return fmt.Errorf("failed to insert rows: %v", err)
	}

	entry, err := lexicalSearch(ctx, conn, "Test Entry", 1)
	if err != nil {
		return fmt.Errorf("failed to query row: %v", err)
	}

	if entry == nil {
		return fmt.Errorf("query result is nil")
	} else if len(entry.Entries) < 1 {
		return fmt.Errorf("query result has no index")
	}

	outputEntry := entry.Entries[0]
	expectedEntry := index.Entries[0]

	if expectedEntry.Name != outputEntry.Name {
		return fmt.Errorf("expected name %s, got %s", expectedEntry.Name, outputEntry.Name)
	}

	if expectedEntry.URI != outputEntry.URI {
		return fmt.Errorf("expected URI %s, got %s", expectedEntry.URI, outputEntry.URI)
	}

	if expectedEntry.DataType != outputEntry.DataType {
		return fmt.Errorf("expected DataType %v, got %v", expectedEntry.DataType, outputEntry.DataType)
	}

	if expectedEntry.SourceType != outputEntry.SourceType {
		return fmt.Errorf("expected SourceType %v, got %v", expectedEntry.SourceType, outputEntry.SourceType)
	}

	if expectedEntry.Content != outputEntry.Content {
		return fmt.Errorf("expected Content %s, got %s", expectedEntry.Content, outputEntry.Content)
	}

	if !reflect.DeepEqual(expectedEntry.WordOccurrences, outputEntry.WordOccurrences) {
		return fmt.Errorf("expected WordOccurrences %v, got %v", expectedEntry.WordOccurrences, outputEntry.WordOccurrences)
	}

	// Delete the test entry
	_, err = conn.Exec(ctx, "DELETE FROM index_list WHERE name = $1", "Test Entry")
	if err != nil {
		return fmt.Errorf("failed to delete test entry: %v", err)
	}

	return nil
}
