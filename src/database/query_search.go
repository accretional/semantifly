package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func searchIndexForTopMatches(ctx context.Context, conn PgxIface, query string, k int) (*pb.Index, error) {
	sqlQuery := `
        WITH query_words AS (
            SELECT word FROM unnest(string_to_array(lower($1), ' ')) AS word
        )
        SELECT 
            name, uri, data_type, source_type, first_added_time, last_refreshed_time, content, word_occurrences,
            ts_rank(search_vector, plainto_tsquery('english', $1)) * 
            (
                SELECT COALESCE(SUM((word_occurrences->word)::int), 0)
                FROM query_words
                WHERE word_occurrences ? word
            ) AS score
        FROM mv_index_list
        WHERE search_vector @@ plainto_tsquery('english', $1)
        ORDER BY score DESC
		LIMIT $2
    `

	rows, err := conn.Query(ctx, sqlQuery, query, k)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	topResults := &pb.Index{
		Entries: make([]*pb.IndexListEntry, 0, k),
	}

	for rows.Next() {
		var entry pb.IndexListEntry
		var dataType, sourceType string
		var firstAddedTime, lastRefreshedTime time.Time
		var wordOccurrencesJSON []byte
		var score float64

		err := rows.Scan(&entry.Name, &entry.URI, &dataType, &sourceType, &firstAddedTime, &lastRefreshedTime, &entry.Content, &wordOccurrencesJSON, &score)
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

		topResults.Entries = append(topResults.Entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return topResults, nil
}
