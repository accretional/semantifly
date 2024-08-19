package subcommands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
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
				Name:            "file1.txt",
				WordOccurrences: map[string]int32{"test": 5, "search": 3, "hill": 2},
			},
			{
				Name:            "file2.txt",
				WordOccurrences: map[string]int32{"test": 2, "search": 7},
			},
			{
				Name:            "file3.txt",
				WordOccurrences: map[string]int32{"other": 1},
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
		args       LexicalSearchArgs
		wantLen    int
		wantFirst  string
		wantOccurs int32
	}{
		{
			name:       "Search for 'test'",
			args:       LexicalSearchArgs{IndexPath: tempDir, SearchTerm: "test", TopN: 2},
			wantLen:    2,
			wantFirst:  "file1.txt",
			wantOccurs: 5,
		},
		{
			name:       "Search for 'search'",
			args:       LexicalSearchArgs{IndexPath: tempDir, SearchTerm: "search", TopN: 1},
			wantLen:    1,
			wantFirst:  "file2.txt",
			wantOccurs: 7,
		},
		{
			name:    "Search for non-existent term",
			args:    LexicalSearchArgs{IndexPath: tempDir, SearchTerm: "nonexistent", TopN: 10},
			wantLen: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := LexicalSearch(tc.args)
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

	args := LexicalSearchArgs{
		IndexPath:  tempDir,
		SearchTerm: "test",
		TopN:       10,
	}

	_, err = LexicalSearch(args)
	if err == nil {
		t.Error("Expected an error for non-existent index file, but got nil")
	}
}

func TestPrintSearchResults(t *testing.T) {
	results := []FileOccurrence{
		{
			FileName:   "file1.txt",
			Occurrence: 5,
		},
		{
			FileName:   "file2.txt",
			Occurrence: 3,
		},
	}

	// capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintSearchResults(results)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedOutput := "File: file1.txt\nOccurrences: 5\n\nFile: file2.txt\nOccurrences: 3\n\n"

	if output != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}
