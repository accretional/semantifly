package subcommands

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strings"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

type LexicalSearchArgs struct {
	IndexPath  string
	SearchTerm string
	TopN       int
}

type fileOccurrence struct {
	FileName   string
	Occurrence int32
}

type occurrenceList []fileOccurrence
type searchMap map[string]occurrenceList // Search Map maps search terms to TermMaps

func createSearchDictionary(ile *pb.IndexListEntry) error {
	srcFile, err := os.Open(ile.URI)

	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
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

// LexicalSearch performs a search in the index for the specified term and returns the top N results ranked by the frequency of the term.
func LexicalSearch(args LexicalSearchArgs) ([]fileOccurrence, error) {
	if args.TopN <= 0 {
		return nil, fmt.Errorf("topn: %d is an invalid amount", args.TopN)
	}
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

	newSearchMap := make(searchMap)
	for _, entry := range index.Entries {
		for word, occ := range entry.WordOccurrences {
			newFileOcc := fileOccurrence{
				FileName:   entry.Name,
				Occurrence: occ,
			}
			if val, ok := newSearchMap[word]; ok {
				newSearchMap[word] = append(val, newFileOcc)
			} else {
				newOccList := []fileOccurrence{newFileOcc}
				newSearchMap[word] = newOccList
			}
		}
	}

	for _, occList := range newSearchMap {
		sort.Slice(occList, func(i, j int) bool {
			return occList[i].Occurrence > occList[j].Occurrence
		})
	}

	resultsLen := len(newSearchMap[args.SearchTerm])
	amountWanted := int(math.Min(float64(args.TopN), float64(resultsLen)))
	return newSearchMap[args.SearchTerm][:amountWanted], nil
}

func PrintSearchResults(results []fileOccurrence) {
	for _, result := range results {
		fmt.Printf("File: %s\nOccurrences: %d\n\n", result.FileName, result.Occurrence)
	}
}
