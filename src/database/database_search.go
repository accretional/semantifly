package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// lexicalSearch performs a lexical search on the index_list table using the provided query.
// It returns a slice of IndexListEntry pointers matching the search criteria, ordered by relevance.
//
// Parameters:
//   - ctx: The context for database operations.
//   - conn: A database connection implementing the PgxIface interface.
//   - query: The search query string.
//   - limit: The maximum number of results to return.
//
// Returns:
//   - A slice of *pb.IndexListEntry containing the search results.
func lexicalSearch(ctx context.Context, conn PgxIface, query string, limit int) (*pb.Index, error) {
    rows, err := conn.Query(ctx, `
        SELECT name, uri, data_type, source_type, first_added_time, last_refreshed_time, content, word_occurrences
        FROM index_list
        WHERE search_vector @@ plainto_tsquery('english', $1)
        ORDER BY ts_rank(search_vector, plainto_tsquery('english', $1)) DESC
        LIMIT $2
    `, query, limit)
    if err != nil {
        return nil, fmt.Errorf("search query failed: %w", err)
    }
    defer rows.Close()

    var results []*pb.IndexListEntry
    for rows.Next() {
        var entry pb.IndexListEntry
        var dataType, sourceType string
        var firstAddedTime, lastRefreshedTime time.Time
        var wordOccurrencesJSON []byte

        err := rows.Scan(
            &entry.Name, &entry.URI, &dataType, &sourceType, &firstAddedTime, &lastRefreshedTime, &entry.Content, &wordOccurrencesJSON)
        if err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        entry.DataType = pb.DataType(pb.DataType_value[dataType])
        entry.SourceType = pb.SourceType(pb.SourceType_value[sourceType])
        entry.FirstAddedTime = timestamppb.New(firstAddedTime)
        entry.LastRefreshedTime = timestamppb.New(lastRefreshedTime)

        err = json.Unmarshal(wordOccurrencesJSON, &entry.WordOccurrences)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal word occurrences: %w", err)
        }

        results = append(results, &entry)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    return &pb.Index{Entries: results}, nil
}

// semanticSearch performs a semantic search on the index_list table using the provided query.
// It generates an embedding for the query and finds the most similar entries in the database.
//
// Parameters:
//   - ctx: The context for the database operation.
//   - conn: A database connection implementing the PgxIface interface.
//   - query: The search query string.
//   - limit: The maximum number of results to return.
//
// Returns:
//   - A slice of *pb.IndexListEntry containing the search results.
func semanticSearch(ctx context.Context, conn PgxIface, query string, limit int) (*pb.Index, error) {

    generator, err := NewEmbeddingGenerator()
    if err != nil {
        return nil, fmt.Errorf("failed to create embedding generator: %w", err)
    }

    queryEmbedding, err := generator.GenerateEmbedding(query)
    if err != nil {
        return nil, fmt.Errorf("failed to generate query embedding: %w", err)
    }

    queryEmbeddingStr := fmt.Sprintf("[%s]", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(queryEmbedding)), ","), "[]"))

    rows, err := conn.Query(ctx, `
        SELECT name, uri, data_type, source_type, first_added_time, last_refreshed_time, content, word_occurrences
        FROM index_list
        ORDER BY embedding <-> $1
        LIMIT $2
    `, queryEmbeddingStr, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to execute semantic search query: %w", err)
    }
    defer rows.Close()

    var results []*pb.IndexListEntry
    for rows.Next() {
        var entry pb.IndexListEntry
        var dataType, sourceType string
        var firstAddedTime, lastRefreshedTime time.Time
        var wordOccurrencesJSON []byte

        err := rows.Scan(
            &entry.Name, &entry.URI, &dataType, &sourceType, &firstAddedTime, &lastRefreshedTime, &entry.Content, &wordOccurrencesJSON)
        if err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        entry.DataType = pb.DataType(pb.DataType_value[dataType])
        entry.SourceType = pb.SourceType(pb.SourceType_value[sourceType])
        entry.FirstAddedTime = timestamppb.New(firstAddedTime)
        entry.LastRefreshedTime = timestamppb.New(lastRefreshedTime)

        err = json.Unmarshal(wordOccurrencesJSON, &entry.WordOccurrences)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal word occurrences: %w", err)
        }

        results = append(results, &entry)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    return &pb.Index{Entries: results}, nil
}