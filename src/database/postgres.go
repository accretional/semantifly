package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Close(context.Context) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

func initializeDatabaseSchema(ctx context.Context, conn PgxIface) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create the main table
	// TODO: Use tsvector instead for lexical search
	_, err = tx.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS index_list (
            name TEXT PRIMARY KEY,
            uri TEXT,
            data_type TEXT,
            source_type TEXT,
            first_added_time TIMESTAMP,
            last_refreshed_time TIMESTAMP,
            content TEXT,
            word_occurrences JSONB
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	// TODO: Create an index for the tsvector once it is added to the table
	_, err = tx.Exec(ctx, `
        CREATE INDEX IF NOT EXISTS idx_word_occurrences ON index_list USING GIN (word_occurrences);
    `)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create materialized view
	_, err = tx.Exec(ctx, `
        CREATE MATERIALIZED VIEW IF NOT EXISTS mv_index_list AS
        SELECT * FROM index_list
        WITH DATA;
        CREATE UNIQUE INDEX IF NOT EXISTS mv_index_list_name_idx ON mv_index_list (name);
    `)
	if err != nil {
		return fmt.Errorf("failed to create materialized view: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// insertRows inserts or updates rows in the index_list table and refreshes the materialized view.
//
// Parameters:
//   - ctx: The context for database operations.
//   - conn: A PgxIface interface for database connection.
//   - index: A pointer to a pb.Index struct containing the entries to be inserted or updated.
func insertRows(ctx context.Context, conn PgxIface, upsertIndex *pb.Index) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, entry := range upsertIndex.Entries {
		wordOccurrencesJSON, err := json.Marshal(entry.WordOccurrences)
		if err != nil {
			return fmt.Errorf("failed to marshal word occurrences: %w", err)
		}

		batch.Queue(`
            INSERT INTO index_list(
                name, uri, data_type, source_type, first_added_time, 
                last_refreshed_time, content, word_occurrences
            )
            VALUES($1, $2, $3, $4, $5, $6, $7, $8)
            ON CONFLICT (name) DO UPDATE SET
                uri = EXCLUDED.uri,
                data_type = EXCLUDED.data_type,
                source_type = EXCLUDED.source_type,
                last_refreshed_time = EXCLUDED.last_refreshed_time,
                content = EXCLUDED.content,
                word_occurrences = EXCLUDED.word_occurrences
        `, entry.Name, entry.URI, entry.DataType.String(), entry.SourceType.String(),
			entry.FirstAddedTime.AsTime(), entry.LastRefreshedTime.AsTime(),
			entry.Content, wordOccurrencesJSON)
	}

	br := tx.SendBatch(ctx, batch)
	_, err = br.Exec()
	br.Close()
	if err != nil {
		return fmt.Errorf("failed to insert rows: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Refresh the materialized view outside the transaction
	_, err = conn.Exec(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY mv_index_list`)
	if err != nil {
		return fmt.Errorf("failed to refresh materialized view: %w", err)
	}

	return nil
}

// queryRow retrieves a single row from the index_list table based on the provided name.
// It returns a pointer to a pb.IndexListEntry struct containing the row data, or an error if the query fails.
//
// Parameters:
//   - ctx: The context for database operations.
//   - conn: A PgxIface interface for database connection.
//   - name: The name of the index entry to retrieve.
//
// Returns:
//   - *pb.IndexListEntry: A pointer to the retrieved index entry.
func getContentMetadata(ctx context.Context, conn PgxIface, name string) (*pb.IndexListEntry, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(ctx)

	var entry pb.IndexListEntry
	var dataType, sourceType string
	var firstAddedTime, lastRefreshedTime time.Time

	err = tx.QueryRow(ctx, `
		SELECT name, uri, data_type, source_type, first_added_time, last_refreshed_time
		FROM index_list 
		WHERE name=$1
	`, name).Scan(
		&entry.Name, &entry.URI, &dataType, &sourceType, &firstAddedTime, &lastRefreshedTime)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no entry found for name: %s", name)
		}

		return nil, fmt.Errorf("QueryRow failed: %w", err)
	}

	entry.DataType = pb.DataType(pb.DataType_value[dataType])
	entry.SourceType = pb.SourceType(pb.SourceType_value[sourceType])
	entry.FirstAddedTime = timestamppb.New(firstAddedTime)
	entry.LastRefreshedTime = timestamppb.New(lastRefreshedTime)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &entry, nil
}
