package database

import (
	"context"
	"fmt"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/proto"
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

	// Create the main table using protobuf
	_, err = tx.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS index_list (
            name TEXT PRIMARY KEY,
            entry BYTEA
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create the protobuf extension functions
	_, err = tx.Exec(ctx, `
        CREATE OR REPLACE FUNCTION pb_get(data bytea, path text)
        RETURNS text
        AS 'postgres-protobuf', 'pb_get'
        LANGUAGE C STRICT IMMUTABLE;

        CREATE OR REPLACE FUNCTION pb_get_json(data bytea, path text)
        RETURNS jsonb
        AS 'postgres-protobuf', 'pb_get_json'
        LANGUAGE C STRICT IMMUTABLE;
    `)
	if err != nil {
		return fmt.Errorf("failed to create protobuf functions: %w", err)
	}

	// Create indexes
	_, err = tx.Exec(ctx, `
        CREATE INDEX IF NOT EXISTS idx_word_occurrences ON index_list USING GIN ((pb_get_json(data, 'word_occurrences')));
        CREATE INDEX IF NOT EXISTS idx_stemmed_word_occurrences ON index_list USING GIN ((pb_get_json(data, 'stemmed_word_occurrences')));
    `)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
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
		ile, err := proto.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf: %w", err)
		}

		batch.Queue(`
            INSERT INTO index_list(name, data)
            VALUES($1, $2)
            ON CONFLICT (name) DO UPDATE SET
                data = EXCLUDED.data
        `, entry.Name, ile)
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

	return nil
}

// getContentMetadata retrieves a single row from the index_list table based on the provided name.
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

	err = tx.QueryRow(ctx, `
        SELECT name, pb_get(entry, 'uri') as uri, pb_get(entry, 'data_type') as data_type, pb_get(entry, 'source_type') as source_type
        FROM index_list 
        WHERE name=$1
    `, name).Scan(&entry.Name, &entry.URI, &dataType, &sourceType)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no entry found for name: %s", name)
		}
		return nil, fmt.Errorf("QueryRow failed: %w", err)
	}

	entry.DataType = pb.DataType(pb.DataType_value[dataType])
	entry.SourceType = pb.SourceType(pb.SourceType_value[sourceType])

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &entry, nil
}
