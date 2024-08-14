package subcommands

import (
	"fmt"
	"os"
	"path"
	"sort"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

type LexicalSearchArgs struct {
	IndexPath  string
	SearchTerm string
	TopN       int
}

type SearchResult struct {
	Entry      *pb.IndexListEntry
	Occurrence int32
}

// LexicalSearch performs a search in the index for the specified term and returns the top N results ranked by the frequency of the term.
func LexicalSearch(args LexicalSearchArgs) ([]SearchResult, error) {
	indexFilePath := path.Join(args.IndexPath, indexFile)
	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index file does not exist at %s", indexFilePath)
		}
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	searchResults := make([]SearchResult, 0)

	for _, entry := range index.Entries {
		if occurrence, found := entry.WordOccurrences[args.SearchTerm]; found {
			searchResults = append(searchResults, SearchResult{
				Entry:      entry,
				Occurrence: occurrence,
			})
		}
	}

	// Sort the search results by occurrence in descending order
	sort.Slice(searchResults, func(i, j int) bool {
		return searchResults[i].Occurrence > searchResults[j].Occurrence
	})

	// If TopN is specified and less than the total results, truncate the results
	if args.TopN > 0 && len(searchResults) > args.TopN {
		searchResults = searchResults[:args.TopN]
	}

	return searchResults, nil
}

func PrintSearchResults(results []SearchResult) {
	for _, result := range results {
		fmt.Printf("File: %s\nOccurrences: %d\n\n", result.Entry.Name, result.Occurrence)
	}
}
