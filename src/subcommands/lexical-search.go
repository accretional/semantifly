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

type FileOccurrence struct {
	FileName   string
	Occurrence int32
}

type OccurrenceList []FileOccurrence

type SearchMap map[string]OccurrenceList // Search Map maps search terms to TermMaps

// LexicalSearch performs a search in the index for the specified term and returns the top N results ranked by the frequency of the term.
func LexicalSearch(args LexicalSearchArgs) ([]FileOccurrence, error) {
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

	searchMap := make(SearchMap)
	for _, entry := range index.Entries {
		for word, occ := range entry.WordOccurrences {
			newFileOcc := FileOccurrence{
				FileName:   entry.Name,
				Occurrence: occ,
			}
			if val, ok := searchMap[word]; ok {
				searchMap[word] = append(val, newFileOcc)
			} else {
				newOccList := []FileOccurrence{newFileOcc}
				searchMap[word] = newOccList
			}
		}
	}

	for _, occList := range searchMap {
		sort.Slice(occList, func(i, j int) bool {
			return occList[i].Occurrence > occList[j].Occurrence
		})
	}

	fmt.Println("searchMap", searchMap)

	searchResults := searchMap[args.SearchTerm]

	fmt.Println("searchResults", searchResults)
	/*
		searchResults := make([]SearchResult, 0)

		for _, entry := range index.Entries {
			if occurrence, found := entry.WordOccurrences[args.SearchTerm]; found {
				searchResults = append(searchResults, SearchResult{
					Entry:      entry,
					Occurrence: occurrence,
				})
			}
		}*/

	// If TopN is specified and less than the total results, truncate the results
	if args.TopN > 0 && len(searchResults) > args.TopN {
		searchResults = searchResults[:args.TopN]
	}

	return searchResults, nil
}

func PrintSearchResults(results []FileOccurrence) {
	for _, result := range results {
		fmt.Printf("File: %s\nOccurrences: %d\n\n", result.FileName, result.Occurrence)
	}
}
