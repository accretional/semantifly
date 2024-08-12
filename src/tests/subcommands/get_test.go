package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGet(t *testing.T) {
	// Set up temporary directory
	indexDir, err := os.MkdirTemp("", "test_get")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcContent1 := "Test Content 1"
	srcContent2 := "Test Content 2"

	srcFile1 := createTempFile(t, indexDir, srcContent1)
	srcFile2 := createTempFile(t, indexDir, srcContent2)

	args_1 := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   true,
		DataURIs:   []string{srcFile1.Name()},
	}

	subcommands.Add(args_1)

	indexPath := createIndexFile(t, tempDir, entries)

	tests := []struct {
		name        string
		fileName    string
		expectedOut string
		expectError bool
	}{
		{"Get existing file without copy", "file1.txt", file1Content, false},
		{"Get existing file with copy", "file2.txt", file2Content, false},
		{"Get non-existent file", "file3.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := Get(GetArgs{
				IndexPath: tempDir,
				FileName:  tt.fileName,
			})

			w.Close()
			os.Stdout = oldStdout
			out, _ := os.ReadAll(r)

			if tt.expectError && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && strings.TrimSpace(string(out)) != tt.expectedOut {
				t.Errorf("Expected output %q, but got %q", tt.expectedOut, strings.TrimSpace(string(out)))
			}
		})
	}
}

func TestGetWithNonExistentIndex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_get_no_index")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = Get(GetArgs{
		IndexPath: tempDir,
		FileName:  "any_file.txt",
	})

	if err == nil {
		t.Errorf("Expected an error due to non-existent index, but got none")
	}
}