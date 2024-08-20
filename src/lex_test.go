package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/bzick/tokenizer"
	"github.com/kljensen/snowball"
)

func TestTokenizer(t *testing.T) {
	parser := tokenizer.New()
	parser.AllowKeywordUnderscore()

	// Read the file content into a string
	fp, err := os.Open("/workspace/semantifly/src/subcommand.go")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer fp.Close()

	fileInfo, err := fp.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)
	_, err = fp.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	fileContent := string(buffer)

	// Parse the entire file content as a string
	stream := parser.ParseString(fileContent).SetHistorySize(10)
	defer stream.Close()

	wordCount := make(map[string]int)

	for stream.IsValid() {
		token := stream.CurrentToken()
		if token.IsNumber() {
			wordCount[token.ValueString()]++
		} else if token.IsKeyword() || token.IsString() {
			stemmedWord, _ := snowball.Stem(token.ValueString(), "english", true)
			wordCount[stemmedWord]++
		}
		stream.GoNext()
	}

	// Output the word count map
	for word, count := range wordCount {
		fmt.Printf("'%s': %d\n", word, count)
	}
}

func TestSnowball(t *testing.T) {
	phrase := `for _, word := range words {
		stemmedWord, _ := snowball.Stem(word, "english", true)
		stemmedWords = append(stemmedWords, stemmedWord)
	}a`

	// Remove punctuation using regex
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	cleanPhrase := re.ReplaceAllString(phrase, " ")

	words := strings.Fields(cleanPhrase)

	// Stem each word
	var stemmedWords []string
	for _, word := range words {
		stemmedWord, _ := snowball.Stem(word, "english", true)
		stemmedWords = append(stemmedWords, stemmedWord)
	}

	fmt.Println(stemmedWords)
}
