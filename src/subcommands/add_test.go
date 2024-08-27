package subcommands

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

func TestAdd(t *testing.T) {
	fmt.Println("--- Testing Add command ---")
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

	// Call the Add function with the buffer
	SubcommandAdd(args)

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

	// Check if the copy of data file for testFilePath was created
	copiesDir := path.Join(tempDir, addedCopiesSubDir)
	if _, err := os.Stat(path.Join(copiesDir, testFilePath)); os.IsNotExist(err) {
		t.Errorf("Data file for %s was not copied", testFilePath)
	}
}

func TestAdd_MultipleFilesSamePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the test files
	testFilePath1 := path.Join(tempDir, "test_file.txt")
	err = os.WriteFile(testFilePath1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	testFilePath2 := path.Join(tempDir, "test_file.txt")

	// Set up test arguments
	args := &pb.AddRequest{
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{testFilePath1, testFilePath2},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the Add function with the buffer
	SubcommandAdd(args)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Checking if the second entry was skipped
	if !strings.Contains(output, "Skipping without refresh") {
		t.Errorf("Expected output 'Skipping without refresh', but got '%s'", output)
	}

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

	// Check if there is only one entry in the index file
	if len(indexMap) > 1 {
		t.Errorf("More than one entry found in the index file\n")
	}
}

func TestAdd_Webpage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the test files
	testWebpageURL := "http://echo.jsontest.com/title/lorem/content/ipsum"

	// Set up test arguments
	args := &pb.AddRequest{
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "webpage",
		MakeCopy:   true,
		DataUris:   []string{testWebpageURL},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the Add function with the buffer
	SubcommandAdd(args)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

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
	if entry, exists := indexMap[testWebpageURL]; !exists {
		t.Errorf("Test file was not added to the index")
	} else {
		// Verify the entry details
		if entry.Name != testWebpageURL {
			t.Errorf("Expected Name %s, got %s", testWebpageURL, entry.Name)
		}
		if entry.URI != testWebpageURL {
			t.Errorf("Expected URI %s, got %s", testWebpageURL, entry.URI)
		}
		if entry.DataType != pb.DataType_TEXT {
			t.Errorf("Expected DataType %v, got %v", pb.DataType_TEXT, entry.DataType)
		}
		if entry.SourceType != pb.SourceType_WEBPAGE {
			t.Errorf("Expected SourceType %v, got %v", pb.SourceType_LOCAL_FILE, entry.SourceType)
		}
		if entry.FirstAddedTime == nil {
			t.Errorf("FirstAddedTime is nil")
		}
	}

	// Check if the copy of data file for testFilePath was created
	copiedWebpagePath := path.Join(tempDir, addedCopiesSubDir, testWebpageURL)
	if _, err := os.Stat(copiedWebpagePath); os.IsNotExist(err) {
		t.Errorf("Data file for %s was not copied", testWebpageURL)
	}

	// Get the content of the copy file
	data, err := os.ReadFile(copiedWebpagePath)
	if err != nil {
		t.Errorf("failed to read file %s: %v", copiedWebpagePath, err)
	}

	ile := &pb.IndexListEntry{}
	err = proto.Unmarshal(data, ile)
	if err != nil {
		t.Errorf("failed to unmarshal IndexListEntry: %v", err)
	}

	// Fetching webpage content
	resp, err := http.Get(testWebpageURL)
	if err != nil {
		t.Errorf("failed to fetch web page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("web page returned non-OK status: %s", resp.Status)
	}

	webpageContent, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read web page content: %v", err)
	}

	// Validating the contents of the copy file
	if ile.Content != string(webpageContent) {
		t.Errorf("Failed to validate webpage copy: Expected \"%s\", got \"%s\"", webpageContent, ile.Content)
	}

	// Check the output in the buffer
	if !strings.Contains(output, "added successfully") {
		t.Errorf("Expected output to contain 'added successfully', but got '%s'", output)
	}
}
