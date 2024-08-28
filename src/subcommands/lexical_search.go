package subcommands

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"

	"github.com/kljensen/snowball"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

type fileOccurrence struct {
	FileName   string
	Occurrence int32
}

type occurrenceList []fileOccurrence
type searchMap map[string]occurrenceList // Search Map maps search terms to the list of their occurrences in files

// LexicalSearch performs a search in the index for the specified term and returns the top N results ranked by the frequency of the term.
func SubcommandLexicalSearch(args *pb.LexicalSearchRequest, indexPath string, w io.Writer) ([]fileOccurrence, error) {
	if args.TopN <= 0 {
		return nil, fmt.Errorf("topn: %d is an invalid amount", args.TopN)
	}
	indexFilePath := path.Join(indexPath, indexFile)
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

	newStemmedSearchMap := make(searchMap)
	for _, entry := range index.Entries {
		for word, occ := range entry.StemmedWordOccurrences {
			newFileOcc := fileOccurrence{
				FileName:   entry.Name,
				Occurrence: occ,
			}
			if val, ok := newStemmedSearchMap[word]; ok {
				newStemmedSearchMap[word] = append(val, newFileOcc)
			} else {
				newOccList := []fileOccurrence{newFileOcc}
				newStemmedSearchMap[word] = newOccList
			}
		}
	}

	term := args.SearchTerm
	stemmedTerm, err := snowball.Stem(term, "english", true)
	if err != nil {
		return nil, fmt.Errorf("failed to stem word: %w", err)
	}

	stemmedResults := newStemmedSearchMap[stemmedTerm]
	nonStemResults := newSearchMap[term]

	// combine stemmed and non-stemmed occurrence numbers to help prioritize files with exact matches
	resultMap := make(map[string]int32)
	for _, result := range stemmedResults {
		resultMap[result.FileName] += result.Occurrence
	}
	for _, result := range nonStemResults {
		resultMap[result.FileName] += result.Occurrence
	}

	var combinedResults []fileOccurrence
	for fileName, occurrence := range resultMap {
		combinedResults = append(combinedResults, fileOccurrence{
			FileName:   fileName,
			Occurrence: occurrence,
		})
	}

	// sort result by descending occurence
	sort.Slice(combinedResults, func(i, j int) bool {
		return combinedResults[i].Occurrence > combinedResults[j].Occurrence
	})

	if len(combinedResults) > int(args.TopN) {
		combinedResults = combinedResults[:args.TopN]
	}

	PrintSearchResults(combinedResults, w)

	return combinedResults, nil
}

func PrintSearchResults(results []fileOccurrence, w io.Writer) {
	for _, result := range results {
		fmt.Fprintf(w, "File: %s\nOccurrences: %d\n\n", result.FileName, result.Occurrence)
	}
}
