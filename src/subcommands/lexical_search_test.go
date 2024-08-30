package subcommands

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

func TestLexicalSearch(t *testing.T) {
	fmt.Println("--- Testing LexicalSearch command ---")
	tempDir, err := os.MkdirTemp("", "lexical_search_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexFilePath := path.Join(tempDir, indexFile)
	mockIndex := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name:                   "file1.txt",
				WordOccurrences:        map[string]int32{"test": 2, "search": 3, "searching": 2},
				StemmedWordOccurrences: map[string]int32{"test": 2, "search": 5},
			},
			{
				Name:                   "file2.txt",
				WordOccurrences:        map[string]int32{"test": 5, "search": 5},
				StemmedWordOccurrences: map[string]int32{"test": 5, "search": 5},
			},
			{
				Name:                   "file3.txt",
				WordOccurrences:        map[string]int32{"other": 1},
				StemmedWordOccurrences: map[string]int32{"other": 1},
			},
		},
	}

	indexData, err := proto.Marshal(mockIndex)
	if err != nil {
		t.Fatalf("Failed to marshal mock index: %v", err)
	}

	err = os.WriteFile(indexFilePath, indexData, 0644)
	if err != nil {
		t.Fatalf("Failed to write mock index file: %v", err)
	}

	testCases := []struct {
		name       string
		args       *pb.LexicalSearchRequest
		wantLen    int
		wantFirst  string
		wantOccurs int32
	}{
		{
			name:       "Search for 'test'",
			args:       &pb.LexicalSearchRequest{SearchTerm: "test", TopN: 2},
			wantLen:    2,
			wantFirst:  "file2.txt",
			wantOccurs: 10,
		},
		{
			name:       "Search for 'search'",
			args:       &pb.LexicalSearchRequest{SearchTerm: "searching", TopN: 1},
			wantLen:    1,
			wantFirst:  "file1.txt",
			wantOccurs: 7,
		},
		{
			name:    "Search for non-existent term",
			args:    &pb.LexicalSearchRequest{SearchTerm: "nonexistent", TopN: 10},
			wantLen: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			results, err := SubcommandLexicalSearch(tc.args, tempDir, &buf)
			if err != nil {
				t.Fatalf("LexicalSearch failed: %v", err)
			}

			if len(results) != tc.wantLen {
				t.Errorf("Expected %d results, got %d", tc.wantLen, len(results))
			}

			if tc.wantLen > 0 {
				if results[0].FileName != tc.wantFirst {
					t.Errorf("Expected first result to be %s, got %s", tc.wantFirst, results[0].FileName)
				}
				if results[0].Occurrence != tc.wantOccurs {
					t.Errorf("Expected first result to have %d occurrences, got %d", tc.wantOccurs, results[0].Occurrence)
				}
			}
		})
	}
}

func TestLexicalSearch_NonExistentIndex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lexical_search_test_nonexistent")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	args := &pb.LexicalSearchRequest{
		SearchTerm: "test",
		TopN:       10,
	}

	var buf bytes.Buffer
	_, err = SubcommandLexicalSearch(args, tempDir, &buf)
	if err == nil {
		t.Error("Expected an error for non-existent index file, but got nil")
	}
}

func TestLexicalSearch_UnexpectedTopN(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lexical_search_test_topn_error")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexFilePath := path.Join(tempDir, indexFile)
	mockIndex := &pb.Index{
		Entries: []*pb.IndexListEntry{
			{
				Name:            "file1.txt",
				WordOccurrences: map[string]int32{"test": 5, "search": 3, "hill": 2},
			},
		},
	}

	indexData, err := proto.Marshal(mockIndex)
	if err != nil {
		t.Fatalf("Failed to marshal mock index: %v", err)
	}

	err = os.WriteFile(indexFilePath, indexData, 0644)
	if err != nil {
		t.Fatalf("Failed to write mock index file: %v", err)
	}

	args := &pb.LexicalSearchRequest{
		SearchTerm: "test",
		TopN:       -4,
	}

	expectedErrorMsg := "topn: -4 is an invalid amount"
	var buf bytes.Buffer
	_, err = SubcommandLexicalSearch(args, tempDir, &buf)
	if err == nil {
		t.Error("Expected an error for bad topN, but got nil")
	} else if strings.Compare(err.Error(), expectedErrorMsg) != 0 {
		t.Errorf("Expected error:\n%s\nGot:\n%s", expectedErrorMsg, err.Error())
	}
}

func TestPrintSearchResults(t *testing.T) {
	results := []fileOccurrence{
		{
			FileName:   "file1.txt",
			Occurrence: 5,
		},
		{
			FileName:   "file2.txt",
			Occurrence: 3,
		},
	}

	var buf bytes.Buffer
	PrintSearchResults(results, &buf)

	output := buf.String()
	expectedOutput := "File: file1.txt\nOccurrences: 5\n\nFile: file2.txt\nOccurrences: 3\n\n"

	if output != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}
