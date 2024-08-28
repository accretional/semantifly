package database

import (
	"context"
	"fmt"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/encoding/protojson"
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
			entry JSONB
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	_, err = tx.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_word_occurrences ON index_list USING GIN ((entry->'wordOccurrences'));
		CREATE INDEX IF NOT EXISTS idx_stemmed_word_occurrences ON index_list USING GIN ((entry->'stemmedWordOccurrences'));
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

func insertRows(ctx context.Context, conn PgxIface, upsertIndex *pb.Index) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, ile := range upsertIndex.Entries {
		ileJson, err := protojson.Marshal(ile)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf to JSON: %w", err)
		}

		batch.Queue(`
            INSERT INTO index_list(name, entry)
            VALUES($1, $2)
            ON CONFLICT (name) DO UPDATE SET
                entry = EXCLUDED.entry
        `, ile.Name, ileJson)
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

func getContentMetadata(ctx context.Context, conn PgxIface, name string) (*pb.ContentMetadata, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(ctx)

	var jsonData []byte
	err = tx.QueryRow(ctx, `
        SELECT entry->'contentMetadata'
        FROM index_list 
        WHERE name=$1
    `, name).Scan(&jsonData)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no entry found for name: %s", name)
		}
		return nil, fmt.Errorf("query row failed: %w", err)
	}

	var contentMetadata pb.ContentMetadata
	err = protojson.Unmarshal(jsonData, &contentMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal content metadata JSON to protobuf: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &contentMetadata, nil
}
