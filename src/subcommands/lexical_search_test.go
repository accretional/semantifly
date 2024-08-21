package subcommands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
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

	args := LexicalSearchArgs{
		IndexPath:  tempDir,
		SearchTerm: "test",
		TopN:       -4,
	}

	expectedErrorMsg := "topn: -4 is an invalid amount"
	_, err = LexicalSearch(args)
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

func TestBuildDictionary(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    map[string]int32
		expectError bool
	}{
		{
			name:  "Basic sentence",
			input: "The quick brown fox jumps over the lazy dog",
			expected: map[string]int32{
				"the":   2,
				"quick": 1,
				"brown": 1,
				"fox":   1,
				"jump":  1,
				"over":  1,
				"lazi":  1,
				"dog":   1,
			},
			expectError: false,
		},
		{
			name:  "Repeated words",
			input: "apple apple banana cherry banana",
			expected: map[string]int32{
				"appl":   2,
				"banana": 2,
				"cherri": 1,
			},
			expectError: false,
		},
		{
			name:  "Numbers and punctuation",
			input: "Hello, world! 123 test 456 test.",
			expected: map[string]int32{
				"hello": 1,
				"world": 1,
				"123":   1,
				"test":  2,
				"456":   1,
			},
			expectError: false,
		},
		{
			name:  "Stemming correctness",
			input: "have eating() having had eat hav:",
			expected: map[string]int32{
				"eat":  2,
				"had":  1,
				"hav":  1,
				"have": 2,
			},
			expectError: false,
		},
		{
			name:        "Empty string",
			input:       "",
			expected:    map[string]int32{},
			expectError: false,
		},
		{
			name:        "Special characters only",
			input:       "!@#$%^&*()+",
			expected:    map[string]int32{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildDictionary(&tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("buildDictionary() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if tt.expected == nil && result == nil {
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("buildDictionary() = %v, want %v", result, tt.expected)
			}
		})
	}
}
