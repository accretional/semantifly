package search

import (
	"fmt"

	"github.com/bzick/tokenizer"
	"github.com/kljensen/snowball"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

// createSearchDictionary processes an IndexListEntry by fetching content from its source,
// tokenizing the content, and populating the WordOccurrences map with word frequencies.
// It takes a pointer to an IndexListEntry as input and returns an error if any issues occur
// during content fetching or processing.

func CreateSearchDictionary(ile *pb.IndexListEntry) error {
	content, err := fetch.FetchFromSource(ile.ContentMetadata.SourceType, ile.ContentMetadata.URI)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}

	if len(content) == 0 {
		return nil
	}

	fileContent := string(content)

	ile.WordOccurrences = make(map[string]int32)
	ile.StemmedWordOccurrences = make(map[string]int32)

	parser := tokenizer.New()
	parser.AllowKeywordUnderscore()

	// parse and tokenize file content
	stream := parser.ParseString(fileContent)
	defer stream.Close()

	for stream.IsValid() {
		token := stream.CurrentToken()
		if token.IsNumber() {
			ile.WordOccurrences[token.ValueString()]++
			ile.StemmedWordOccurrences[token.ValueString()]++
		} else if token.IsKeyword() || token.IsString() {
			stemmedWord, err := snowball.Stem(token.ValueString(), "english", true)
			if err != nil {
				return fmt.Errorf("failed to stem word: %w", err)
			}
			ile.WordOccurrences[token.ValueString()]++
			ile.StemmedWordOccurrences[stemmedWord]++
		}
		stream.GoNext()
	}

	return nil
}
