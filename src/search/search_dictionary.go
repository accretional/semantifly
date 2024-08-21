package search

import (
	"fmt"
	"strings"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

// createSearchDictionary processes an IndexListEntry by fetching content from its source,
// tokenizing the content, and populating the WordOccurrences map with word frequencies.
// It takes a pointer to an IndexListEntry as input and returns an error if any issues occur
// during content fetching or processing.
func CreateSearchDictionary(ile *pb.IndexListEntry) error {

	content, err := fetch.FetchFromSource(ile.SourceType, ile.URI)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create and populate the word_occurrences map
	ile.WordOccurrences = make(map[string]int32)
	tokens := strings.Fields(strings.ToLower(string(content)))
	for _, token := range tokens {
		ile.WordOccurrences[token]++
	}

	return nil
}
