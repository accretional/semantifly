package subcommands

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func TestGet(t *testing.T) {
	fmt.Println("--- Testing Get command ---")
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFilePath := path.Join(tempDir, "test_file.txt")
	testContent := "Test content"
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addArgs := &pb.AddRequest{
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{testFilePath},
	}

	var addBuf bytes.Buffer

	err = SubcommandAdd(addArgs, &addBuf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	// Check if the index file was created
	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	getArgs := &pb.GetRequest{
		IndexPath: tempDir,
		Name:      testFilePath,
	}

	var getBuf bytes.Buffer

	_, err = SubcommandGet(getArgs, &getBuf)
	if err != nil {
		t.Fatalf("Get function returned an error: %v", err)
	}

	output := getBuf.String()

	// Get command prints out the content and a new line in the end
	testContent = testContent + "\n"
	if output != testContent {
		t.Errorf("Expected output '%s', but got '%s'", testContent, output)
	}
}

func TestGet_Webpage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testWebpageURL := "http://echo.jsontest.com/title/lorem/content/ipsum"

	addArgs := &pb.AddRequest{
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "webpage",
		MakeCopy:   true,
		DataUris:   []string{testWebpageURL},
	}

	var addBuf bytes.Buffer

	err = SubcommandAdd(addArgs, &addBuf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	getArgs := &pb.GetRequest{
		IndexPath: tempDir,
		Name:      testWebpageURL,
	}

	var getBuf bytes.Buffer

	_, err = SubcommandGet(getArgs, &getBuf)
	if err != nil {
		t.Fatalf("Get function returned an error: %v", err)
	}

	output := getBuf.String()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(testWebpageURL)
	if err != nil {
		t.Errorf("failed to fetch web page: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("web page returned non-OK status: %s", resp.Status)
		return
	}

	webpageContent, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read web page content: %v", err)
		return
	}

	webpageContentStr := string(webpageContent) + "\n"

	if output != webpageContentStr {
		t.Errorf("Failed to validate webpage copy: Expected \"%s\", got \"%s\"", webpageContent, output)
	}
}
