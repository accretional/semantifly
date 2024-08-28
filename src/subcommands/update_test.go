package subcommands

import (
	"fmt"
	"os"
	"path"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

func TestUpdate(t *testing.T) {
	fmt.Println("--- Testing Update command ---")
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the test files
	testFilePath := path.Join(tempDir, "test_file.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set up test arguments
	args := AddArgs{
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataURIs:   []string{testFilePath},
	}

	// Call the Add function
	Add(args)

	// Check if the index file was created
	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	// Updating the index entry
	updatedFilePath := path.Join(tempDir, "test_file_updated.txt")
	updatedContent := "test content - updated"
	err = os.WriteFile(updatedFilePath, []byte(updatedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set up test arguments
	updateArgs := UpdateArgs{
		IndexPath:  tempDir,
		Name:       testFilePath,
		UpdateCopy: "true",
		DataURI:    updatedFilePath,
	}

	// Call the Add function
	Update(updateArgs)

	// Read the index file
	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		t.Fatalf("Failed to read index file: %v", err)
	}

	// Check if the test file was updated in the index
	if entry, exists := indexMap[testFilePath]; !exists {
		t.Errorf("Test file was not added to the index")
	} else {
		// Verify the entry details
		if entry.Name != testFilePath {
			t.Errorf("Expected Name %s, got %s", testFilePath, entry.Name)
		}
		if entry.ContentMetadata.URI != updatedFilePath {
			t.Errorf("Expected URI %s, got %s", testFilePath, entry.ContentMetadata.URI)
		}
		if entry.ContentMetadata.DataType != pb.DataType_TEXT {
			t.Errorf("Expected DataType %v, got %v", pb.DataType_TEXT, entry.ContentMetadata.DataType)
		}
		if entry.ContentMetadata.SourceType != pb.SourceType_LOCAL_FILE {
			t.Errorf("Expected SourceType %v, got %v", pb.SourceType_LOCAL_FILE, entry.ContentMetadata.SourceType)
		}
	}

	// Check if the copy of data file for testFilePath was updated
	copiedFile := path.Join(tempDir, addedCopiesSubDir, testFilePath)
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("Data file for %s was not created", updatedFilePath)
	}

	data, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Errorf("failed to read copy file: %v", err)
	}

	ile := &pb.IndexListEntry{}
	err = proto.Unmarshal(data, ile)
	if err != nil {
		t.Errorf("failed to unmarshal IndexListEntry: %v", err)
	}

	if ile.Content != updatedContent {
		t.Errorf("Copy file for %s not updated. Expected content \"%s\", found \"%s\"", testFilePath, updatedContent, ile.Content)
	}
}
