package database

import (
	"context"
	"fmt"
	"strings"

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

func InitializeDatabaseSchema(ctx context.Context, conn PgxIface) error {

	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS index_list (
			name TEXT PRIMARY KEY,
			entry JSONB
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create the table: %w", err)
	}

	return nil

}

func CreateProtoFieldIndex(ctx context.Context, conn PgxIface, fieldName string) error {

	indexName := strings.ReplaceAll(strings.ReplaceAll(fieldName, "->", "_"), "'", "")
	query := `CREATE INDEX IF NOT EXISTS idx_` + indexName + ` ON index_list USING GIN ((` + fieldName + `));`

	_, err := conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

func InsertRows(ctx context.Context, conn PgxIface, upsertIndex *pb.Index) error {
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

func DeleteRows(ctx context.Context, conn PgxIface, names []string) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM index_list 
		WHERE name=ANY($1)
	`, names)
	if err != nil {
		return fmt.Errorf("failed to delete row from table: %w", err)
	}

	return nil
}

func GetContentMetadata(ctx context.Context, conn PgxIface, name string) (*pb.ContentMetadata, error) {
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
