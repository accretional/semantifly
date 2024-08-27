package subcommands

import (
	"fmt"
	"os"
	"path"
	"testing"

	"accretional.com/semantifly/grpcclient"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func TestServerAdd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the test files
	testFilePath := path.Join(tempDir, "test_file.txt")
	err = os.WriteFile(testFilePath, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set up test arguments
	args := &pb.AddRequest{
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{testFilePath},
	}

	res, err := grpcclient.Add(args)
	fmt.Println(res)
}
