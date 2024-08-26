package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type entryWithScore struct {
    entry *pb.IndexListEntry
    score float64
}

func searchIndexByWord(ctx context.Context, conn PgxIface, query string, k int) (*pb.Index, error) {
	entries, err := fetchEntriesByWordFrequency(ctx, conn, query, k)
	if err != nil {
		return nil, err
	}

	return &pb.Index{Entries: entries}, nil
}

func fetchEntriesByWordFrequency(ctx context.Context, conn PgxIface, word string, k int) ([]*pb.IndexListEntry, error) {
	sqlQuery := `
        SELECT 
            name, uri, data_type, source_type, first_added_time, last_refreshed_time, content, word_occurrences,
            (word_occurrences->$1)::int AS word_frequency
        FROM mv_index_list
        WHERE word_occurrences ? $1
        ORDER BY (word_occurrences->$1)::int DESC
        LIMIT $2
    `

	rows, err := conn.Query(ctx, sqlQuery, word, k)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var entries []*pb.IndexListEntry
	for rows.Next() {
		entry, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry.entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// searchIndexForTopMatches searches the index for entries matching the given query and returns the top k matches.
// It fetches matching entries from the database, scores and sorts them based on relevance to the query,
// and returns the top k entries as an Index protobuf message.
//
// Parameters:
//   - ctx: Context for the database operation
//   - conn: Database connection interface
//   - query: Search query string
//   - k: Number of top matches to return
//
// Returns:
//   - *pb.Index: Protobuf Index message containing the top k matching entries
func searchIndexByPhrase(ctx context.Context, conn PgxIface, query string, k int) (*pb.Index, error) {
    entries, err := fetchMatchingEntries(ctx, conn, query)
    if err != nil {
        return nil, err
    }

    scoredEntries := scoreAndSortEntries(entries, query)
    topEntries := getTopEntries(scoredEntries, k)

    return &pb.Index{Entries: topEntries}, nil
}


// fetchMatchingEntries retrieves entries from the database that match the given query.
// It performs a full-text search using PostgreSQL's ts_rank function.
//
// Parameters:
//   - ctx: The context for database operations.
//   - conn: A PostgreSQL connection interface.
//   - query: The search query string.
//
// Returns:
//   - []entryWithScore: A slice of matching entries with their search rank scores.
func fetchMatchingEntries(ctx context.Context, conn PgxIface, query string) ([]entryWithScore, error) {
    sqlQuery := `
        SELECT 
            name, uri, data_type, source_type, first_added_time, last_refreshed_time, content, word_occurrences,
            ts_rank(search_vector, plainto_tsquery('english', $1)) AS rank
        FROM mv_index_list
        WHERE search_vector @@ plainto_tsquery('english', $1)
    `

    rows, err := conn.Query(ctx, sqlQuery, query)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer rows.Close()

    var entries []entryWithScore
    for rows.Next() {
        entry, err := scanRow(rows)
        if err != nil {
            return nil, err
        }
        entries = append(entries, entry)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %w", err)
    }

    return entries, nil
}

func scanRow(row pgx.Row) (entryWithScore, error) {
    var entry pb.IndexListEntry
    var dataType, sourceType string
    var firstAddedTime, lastRefreshedTime time.Time
    var wordOccurrencesJSON []byte
    var rank float64

    err := row.Scan(&entry.Name, &entry.URI, &dataType, &sourceType,
        &firstAddedTime, &lastRefreshedTime, &entry.Content,
        &wordOccurrencesJSON, &rank)
    if err != nil {
        return entryWithScore{}, fmt.Errorf("failed to scan row: %w", err)
    }

    var wordOccurrences map[string]int
    err = json.Unmarshal(wordOccurrencesJSON, &wordOccurrences)
    if err != nil {
        return entryWithScore{}, fmt.Errorf("failed to unmarshal word occurrences: %w", err)
    }

    entry.DataType = pb.DataType(pb.DataType_value[dataType])
    entry.SourceType = pb.SourceType(pb.SourceType_value[sourceType])
    entry.FirstAddedTime = timestamppb.New(firstAddedTime)
    entry.LastRefreshedTime = timestamppb.New(lastRefreshedTime)
    entry.WordOccurrences = convertToInt32Map(wordOccurrences)

    return entryWithScore{&entry, rank}, nil
}


// scoreAndSortEntries calculates scores for the given entries based on the query
// and sorts them in descending order of their scores.
//
// Parameters:
//   - entries: A slice of entryWithScore to be scored and sorted.
//   - query: A string containing the search query.
//
// Returns:
//   - A slice of entryWithScore sorted by score in descending order.
func scoreAndSortEntries(entries []entryWithScore, query string) []entryWithScore {
    queryWords := strings.Fields(strings.ToLower(query))
    
    for i := range entries {
        entries[i].score = calculateScore(queryWords, entries[i].entry.WordOccurrences, entries[i].score)
    }

    sort.Slice(entries, func(i, j int) bool {
        return entries[i].score > entries[j].score
    })

    return entries
}

func calculateScore(queryWords []string, wordOccurrences map[string]int32, rank float64) float64 {
    var score float64
    for _, word := range queryWords {
        if count, ok := wordOccurrences[word]; ok {
            score += float64(count)
        }
    }
    return score * rank
}


// getTopEntries returns the top k entries from the given slice of scored entries.
// If the input slice has fewer than k elements, all entries are returned.
// The returned slice contains pointers to IndexListEntry objects.
//
// Parameters:
//   - scoredEntries: A slice of entryWithScore objects to be filtered.
//   - k: The maximum number of top entries to return.
//
// Returns:
//   - A slice of pointers to pb.IndexListEntry objects, containing at most k entries.
func getTopEntries(scoredEntries []entryWithScore, k int) []*pb.IndexListEntry {
    if len(scoredEntries) > k {
        scoredEntries = scoredEntries[:k]
    }

    entries := make([]*pb.IndexListEntry, len(scoredEntries))
    for i, e := range scoredEntries {
        entries[i] = e.entry
    }
    return entries
}

func convertToInt32Map(m map[string]int) map[string]int32 {
    result := make(map[string]int32)
    for k, v := range m {
        result[k] = int32(v)
    }
    return result
}