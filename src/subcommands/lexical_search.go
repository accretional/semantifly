package subcommands

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/kljensen/snowball"

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

func tokenizeFile(content string) ([]string, error) {
	// Remove punctuation using regex
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	cleanPhrase := re.ReplaceAllString(content, " ")

	words := strings.Fields(cleanPhrase)

	// Stem each word
	var stemmedWords []string
	for _, word := range words {
		stemmedWord, err := snowball.Stem(word, "english", true)
		if err != nil {
			return nil, fmt.Errorf("failed to stem word: %w", err)
		}
		stemmedWords = append(stemmedWords, stemmedWord)
	}

	return stemmedWords, nil
}

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
	tokens, err := tokenizeFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to tokenize file: %w", err)
	}

	for _, token := range tokens {
		ile.WordOccurrences[token]++
	}

	fmt.Println(ile.WordOccurrences)
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

	stemmedTerm, err := snowball.Stem(args.SearchTerm, "english", true)
	if err != nil {
		return nil, fmt.Errorf("failed to stem word: %w", err)
	}

	resultsLen := len(newSearchMap[stemmedTerm])
	amountWanted := int(math.Min(float64(args.TopN), float64(resultsLen)))
	return newSearchMap[stemmedTerm][:amountWanted], nil
}

func PrintSearchResults(results []fileOccurrence) {
	for _, result := range results {
		fmt.Printf("File: %s\nOccurrences: %d\n\n", result.FileName, result.Occurrence)
	}
}
