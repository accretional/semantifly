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

	var filesData []*pb.ContentMetadata

	testFileData := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath,
	}
	filesData = append(filesData, testFileData)

	addArgs := &pb.AddRequest{
		FilesData: filesData,
		MakeCopy:  true,
	}

	var addBuf bytes.Buffer

	err = SubcommandAdd(addArgs, tempDir, &addBuf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	// Check if the index file was created
	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	getArgs := &pb.GetRequest{
		Name: testFilePath,
	}

	var getBuf bytes.Buffer

	resp, err := SubcommandGet(getArgs, tempDir, &getBuf)
	if err != nil {
		t.Fatalf("Get function returned an error: %v", err)
	}

	// Get command prints out the content and a new line in the end
	if resp != testContent {
		t.Errorf("Expected output '%s', but got '%s'", testContent, resp)
	}
}

func TestGet_Webpage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testWebpageURL := "http://echo.jsontest.com/title/lorem/content/ipsum"

	var filesData []*pb.ContentMetadata

	testWebData := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 1,
		URI:        testWebpageURL,
	}

	filesData = append(filesData, testWebData)

	addArgs := &pb.AddRequest{
		FilesData: filesData,
		MakeCopy:  true,
	}

	var addBuf bytes.Buffer

	err = SubcommandAdd(addArgs, tempDir, &addBuf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	getArgs := &pb.GetRequest{
		Name: testWebpageURL,
	}

	var getBuf bytes.Buffer

	getResp, err := SubcommandGet(getArgs, tempDir, &getBuf)
	if err != nil {
		t.Fatalf("Get function returned an error: %v", err)
	}

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

	webpageContentStr := string(webpageContent)

	if getResp != webpageContentStr {
		t.Errorf("Failed to validate webpage copy: Expected \"%s\", got \"%s\"", webpageContent, getResp)
	}
}
