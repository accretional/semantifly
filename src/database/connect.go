package database

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Close(context.Context) error
}

func establishConnection(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return conn, nil
}

func insertRow(ctx context.Context, conn PgxIface, index *pb.Index) error {

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS index_list (
			name TEXT PRIMARY KEY,
			uri TEXT,
			data_type TEXT,
			source_type TEXT,
			first_added_time TIMESTAMP,
			last_refreshed_time TIMESTAMP,
			content TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	batch := &pgx.Batch{}
	for _, entry := range index.Entries {
		batch.Queue(`
			INSERT INTO index_list(name, uri, data_type, source_type, first_added_time, last_refreshed_time, content)
			VALUES($1, $2, $3, $4, $5, $6, $7)
		`, entry.Name, entry.URI, entry.DataType.String(), entry.SourceType.String(), entry.FirstAddedTime.AsTime(), entry.LastRefreshedTime.AsTime(), entry.Content)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	_, err = br.Exec()
	if err != nil {
		return fmt.Errorf("failed to insert rows: %w", err)
	}

	return nil
}

func queryRow(ctx context.Context, conn PgxIface, name string) (*pb.IndexListEntry, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(ctx)

	var entry pb.IndexListEntry
	var dataType, sourceType string
	var firstAddedTime, lastRefreshedTime time.Time

	err = tx.QueryRow(ctx, `
		SELECT name, uri, data_type, source_type, first_added_time, last_refreshed_time, content 
		FROM index_list 
		WHERE name=$1
	`, name).Scan(
		&entry.Name, &entry.URI, &dataType, &sourceType, &firstAddedTime, &lastRefreshedTime, &entry.Content)
	if err != nil {
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