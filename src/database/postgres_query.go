package database

import (
	"context"
	"fmt"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"github.com/kljensen/snowball"
)

func getTopIndexesByWord(ctx context.Context, conn PgxIface, word string, limit int) (*pb.Index, error) {
    tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %w", err)
    }
    defer tx.Rollback(ctx)

    // Stem the input word
    stemmedWord, err := snowball.Stem(word, "english", true)
    if err != nil {
        return nil, fmt.Errorf("failed to stem word: %w", err)
    }

    query := `
        SELECT 
            name,
            entry->'contentMetadata' AS content_metadata
        FROM index_list
        WHERE entry->'stemmedWordOccurrences' ? $1
        ORDER BY (entry->'stemmedWordOccurrences'->$1)::int DESC
        LIMIT $2
    `

    rows, err := tx.Query(ctx, query, stemmedWord, limit)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()

    result := &pb.Index{}
    for rows.Next() {
        var name string
        var contentMetadataJSON []byte

        if err := rows.Scan(&name, &contentMetadataJSON); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        var entry pb.IndexListEntry
        entry.Name = name

        var contentMetadata pb.ContentMetadata
        if err := protojson.Unmarshal(contentMetadataJSON, &contentMetadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal content metadata: %w", err)
        }
        entry.ContentMetadata = &contentMetadata

        result.Entries = append(result.Entries, &entry)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return result, nil
}

func getTopIndexesForPhrase(ctx context.Context, conn PgxIface, phrase string, limit int) (*pb.Index, error) {
    tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %w", err)
    }
    defer tx.Rollback(ctx)

    // Prepare the query
    query := `
        SELECT 
            name,
            entry->'contentMetadata' AS content_metadata
        FROM index_list
        WHERE search_vector @@ plainto_tsquery('english', $1)
        ORDER BY ts_rank_cd(search_vector, plainto_tsquery('english', $1)) DESC
        LIMIT $2
    `

    // Execute the query
    rows, err := tx.Query(ctx, query, phrase, limit)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()

    result := &pb.Index{}
    for rows.Next() {
        var name string
        var contentMetadataJSON []byte

        if err := rows.Scan(&name, &contentMetadataJSON); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        var entry pb.IndexListEntry
        entry.Name = name

        var contentMetadata pb.ContentMetadata
        if err := protojson.Unmarshal(contentMetadataJSON, &contentMetadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal content metadata: %w", err)
        }
        entry.ContentMetadata = &contentMetadata

        result.Entries = append(result.Entries, &entry)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return result, nil
}