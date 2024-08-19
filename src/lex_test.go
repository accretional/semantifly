package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kljensen/snowball"
)

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
