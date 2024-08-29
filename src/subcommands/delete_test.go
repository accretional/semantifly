package subcommands

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func TestDelete(t *testing.T) {
	fmt.Println("--- Testing Delete command ---")
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_delete")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the test files
	testFilePath1 := path.Join(tempDir, "test_file1.txt")
	err = os.WriteFile(testFilePath1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	testFilePath2 := path.Join(tempDir, "test_file2.txt")
	err = os.WriteFile(testFilePath2, []byte("test content 2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	var filesData []*pb.ContentMetadata

	testFileData1 := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath1,
	}
	testFileData2 := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath2,
	}

	filesData = append(filesData, testFileData1, testFileData2)

	addArgs := &pb.AddRequest{
		FilesData: filesData,
		MakeCopy:  true,
	}

	var addBuf bytes.Buffer

	err = SubcommandAdd(addArgs, tempDir, &addBuf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	// Test case
	deleteArgs := &pb.DeleteRequest{
		DeleteCopy: true,
		DataUris:   []string{testFilePath1},
	}

	var deleteBuf bytes.Buffer
	// Run the Delete function
	err = SubcommandDelete(deleteArgs, tempDir, &deleteBuf)
	if err != nil {
		t.Fatalf("Delete function returned an error: %v", err)
	}

	// Verify the results
	indexFilePath := path.Join(tempDir, indexFile)
	updatedIndex, err := readIndex(indexFilePath, false)
	if err != nil {
		t.Fatalf("Failed to read updated index: %v", err)
	}

	// Check if test file 1 was deleted from the index
	if _, exists := updatedIndex[testFilePath1]; exists {
		t.Errorf("%s was not deleted from the index", testFilePath1)
	}

	// Check if the copy of test file 1 was deleted
	copiesDir := path.Join(tempDir, addedCopiesSubDir)
	if _, err := os.Stat(path.Join(copiesDir, testFilePath1)); !os.IsNotExist(err) {
		t.Errorf("Data file for %s was not deleted", testFilePath1)
	}
}
