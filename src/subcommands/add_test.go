package subcommands

import (
	"os"
	"path"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func TestAdd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFilePath := path.Join(tempDir, "test_file.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set up test arguments
	args := AddArgs{
		IndexPath:  tempDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   false,
		DataURIs:   []string{testFilePath},
	}

	// Call the Add function
	Add(args)

	// Check if the index file was created
	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	// Read the index file
	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		t.Fatalf("Failed to read index file: %v", err)
	}

	// Check if the test file was added to the index
	if entry, exists := indexMap[testFilePath]; !exists {
		t.Errorf("Test file was not added to the index")
	} else {
		// Verify the entry details
		if entry.Name != testFilePath {
			t.Errorf("Expected Name %s, got %s", testFilePath, entry.Name)
		}
		if entry.URI != testFilePath {
			t.Errorf("Expected URI %s, got %s", testFilePath, entry.URI)
		}
		if entry.DataType != pb.DataType_TEXT {
			t.Errorf("Expected DataType %v, got %v", pb.DataType_TEXT, entry.DataType)
		}
		if entry.SourceType != pb.SourceType_LOCAL_FILE {
			t.Errorf("Expected SourceType %v, got %v", pb.SourceType_LOCAL_FILE, entry.SourceType)
		}
		if entry.FirstAddedTime == nil {
			t.Errorf("FirstAddedTime is nil")
		}
	}
}

