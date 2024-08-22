package search

import (
	"fmt"
	"io"
	"os"

	"github.com/bzick/tokenizer"
	"github.com/kljensen/snowball"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

// createSearchDictionary processes an IndexListEntry by fetching content from its source,
// tokenizing the content, and populating the WordOccurrences map with word frequencies.
// It takes a pointer to an IndexListEntry as input and returns an error if any issues occur
// during content fetching or processing.

func buildDictionary(content *string, stem bool) (map[string]int32, error) {
	// check for nil pointer
	wordMap := make(map[string]int32)
	if content == nil {
		return wordMap, nil
	} else if *content == "" {
		return wordMap, nil
	}

	parser := tokenizer.New()
	parser.AllowKeywordUnderscore()
	fileContent := *content

	// parse file content
	stream := parser.ParseString(fileContent)
	defer stream.Close()

	for stream.IsValid() {
		token := stream.CurrentToken()
		if token.IsNumber() {
			wordMap[token.ValueString()]++
		} else if token.IsKeyword() || token.IsString() {
			stemmedWord, err := snowball.Stem(token.ValueString(), "english", true)
			if err != nil {
				return nil, fmt.Errorf("failed to stem word: %w", err)
			}
			wordMap[stemmedWord]++
		}
		stream.GoNext()
	}

	return wordMap, nil
}

func CreateSearchDictionary(ile *pb.IndexListEntry) error {
	srcFile, err := os.Open(ile.URI)

	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	fileContent := string(content)

	// build stemmed dictionary
	ile.WordOccurrences, err = buildDictionary(&fileContent, true)
	if err != nil {
		return fmt.Errorf("failed to build word occurence map: %w", err)
	}
	return nil
}

// func CreateSearchDictionary(ile *pb.IndexListEntry) error {

// 	content, err := fetch.FetchFromSource(ile.SourceType, ile.URI)
// 	if err != nil {
// 		return fmt.Errorf("failed to read source file: %w", err)
// 	}

// 	// Create and populate the word_occurrences map
// 	ile.WordOccurrences = make(map[string]int32)
// 	tokens := strings.Fields(strings.ToLower(string(content)))
// 	for _, token := range tokens {
// 		ile.WordOccurrences[token]++
// 	}

// 	return nil
// }
