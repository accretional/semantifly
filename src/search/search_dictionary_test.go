package search

import (
	"reflect"
	"testing"
)

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
