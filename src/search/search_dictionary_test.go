package search

import (
	"os"
	"reflect"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func TestCreateSearchDictionary(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedStemmed    map[string]int32
		expectedNonStemmed map[string]int32
		expectError        bool
	}{
		{
			name:  "Basic sentence",
			input: "The quick brown fox jumps over the lazy dog",
			expectedStemmed: map[string]int32{
				"the":   2,
				"quick": 1,
				"brown": 1,
				"fox":   1,
				"jump":  1,
				"over":  1,
				"lazi":  1,
				"dog":   1,
			},
			expectedNonStemmed: map[string]int32{
				"The":   1,
				"the":   1,
				"quick": 1,
				"brown": 1,
				"fox":   1,
				"jumps": 1,
				"over":  1,
				"lazy":  1,
				"dog":   1,
			},
			expectError: false,
		},
		{
			name:  "Repeated words",
			input: "apple apple banana cherry banana",
			expectedStemmed: map[string]int32{
				"appl":   2,
				"banana": 2,
				"cherri": 1,
			},
			expectedNonStemmed: map[string]int32{
				"apple":  2,
				"banana": 2,
				"cherry": 1,
			},
			expectError: false,
		},
		{
			name:  "Numbers and punctuation",
			input: "Hello, world! 123 test 456 test.",
			expectedStemmed: map[string]int32{
				"hello": 1,
				"world": 1,
				"123":   1,
				"test":  2,
				"456":   1,
			},
			expectedNonStemmed: map[string]int32{
				"Hello": 1,
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
			expectedStemmed: map[string]int32{
				"eat":  2,
				"had":  1,
				"hav":  1,
				"have": 2,
			},
			expectedNonStemmed: map[string]int32{
				"have":   1,
				"eating": 1,
				"having": 1,
				"had":    1,
				"eat":    1,
				"hav":    1,
			},
			expectError: false,
		},
		{
			name:               "Empty string",
			input:              "",
			expectedStemmed:    nil,
			expectedNonStemmed: nil,
			expectError:        false,
		},
		{
			name:               "Special characters only",
			input:              "!@#$%^&*()+",
			expectedStemmed:    map[string]int32{},
			expectedNonStemmed: map[string]int32{},
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatalf("Failed to create temporary file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			// write test input to file
			if _, err := tempFile.Write([]byte(tt.input)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tempFile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			ile := &pb.IndexListEntry{
				URI: tempFile.Name(),
			}

			err = CreateSearchDictionary(ile)

			if (err != nil) != tt.expectError {
				t.Errorf("CreateSearchDictionary() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !reflect.DeepEqual(ile.StemmedWordOccurrences, tt.expectedStemmed) {
				t.Errorf("CreateSearchDictionary() StemmedWordOccurrences = %v, want %v", ile.StemmedWordOccurrences, tt.expectedStemmed)
			}

			if !reflect.DeepEqual(ile.WordOccurrences, tt.expectedNonStemmed) {
				t.Errorf("CreateSearchDictionary() WordOccurrences = %v, want %v", ile.WordOccurrences, tt.expectedNonStemmed)
			}
		})
	}
}
