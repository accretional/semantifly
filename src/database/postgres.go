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
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func establishConnection(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return conn, nil
}

func insertRows(ctx context.Context, conn PgxIface, index *pb.Index) error {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS index_list (
			name TEXT PRIMARY KEY,
			uri TEXT,
			data_type TEXT,
			source_type TEXT,
			first_added_time TIMESTAMP,
			last_refreshed_time TIMESTAMP,
			content TEXT,
			search_vector tsvector
		)
    `)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create GIN index on search_vector if it doesn't exist
	_, err = tx.Exec(ctx, `
        CREATE INDEX IF NOT EXISTS idx_search_vector ON index_list USING GIN (search_vector);
    `)
	if err != nil {
		return fmt.Errorf("failed to create search_vector index: %w", err)
	}

	batch := &pgx.Batch{}
	for _, entry := range index.Entries {
		if err != nil {
			return fmt.Errorf("failed to marshal word occurrences: %w", err)
		}
		batch.Queue(`
            INSERT INTO index_list(
                name, uri, data_type, source_type, first_added_time, 
                last_refreshed_time, content, search_vector
            )
            VALUES($1, $2, $3, $4, $5, $6, $7, to_tsvector('english', $7))
            ON CONFLICT (name) DO UPDATE SET
                uri = EXCLUDED.uri,
                data_type = EXCLUDED.data_type,
                source_type = EXCLUDED.source_type,
                last_refreshed_time = EXCLUDED.last_refreshed_time,
                content = EXCLUDED.content,
                search_vector = to_tsvector('english', EXCLUDED.content)
        `, entry.Name, entry.URI, entry.DataType.String(), entry.SourceType.String(),
			entry.FirstAddedTime.AsTime(), entry.LastRefreshedTime.AsTime(),
			entry.Content)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	_, err = br.Exec()
	if err != nil {
		return fmt.Errorf("failed to insert rows: %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func getTopMatches(ctx context.Context, conn PgxIface, query string) (*pb.Index, error) {
	sqlQuery := `
        WITH query_words AS (
            SELECT unnest(string_to_array(lower($1), ' ')) AS word
        )
        SELECT 
            name, uri, data_type, source_type, first_added_time, last_refreshed_time, content,
            SUM((SELECT COUNT(*) FROM query_words WHERE content ILIKE '%' || word || '%')) AS score
        FROM index_list
        WHERE search_vector @@ plainto_tsquery('english', $1)
        GROUP BY name, uri, data_type, source_type, first_added_time, last_refreshed_time, content
        ORDER BY score DESC
        LIMIT 3
    `

	rows, err := conn.Query(ctx, sqlQuery, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	topResults := &pb.Index{
		Entries: make([]*pb.IndexListEntry, 0, 3),
	}

	for rows.Next() {
		var entry pb.IndexListEntry
		var dataType, sourceType string
		var firstAddedTime, lastRefreshedTime time.Time
		var score int

		err := rows.Scan(&entry.Name, &entry.URI, &dataType, &sourceType, &firstAddedTime, &lastRefreshedTime, &entry.Content, &score)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		entry.DataType = pb.DataType(pb.DataType_value[dataType])
		entry.SourceType = pb.SourceType(pb.SourceType_value[sourceType])
		entry.FirstAddedTime = timestamppb.New(firstAddedTime)
		entry.LastRefreshedTime = timestamppb.New(lastRefreshedTime)

		topResults.Entries = append(topResults.Entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return topResults, nil
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
