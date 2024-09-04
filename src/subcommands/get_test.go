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
	"time"
)

func TestGet(t *testing.T) {
	fmt.Println("--- Testing Get command ---")
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the test files
	testFilePath := path.Join(tempDir, "test_file.txt")
	testContent := "Test content"
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup database connection
	ctx, conn, err := setupDatabaseForTesting()
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL database: %v", err)
	}
	defer closeTestingDatabase()
	defer conn.Close(ctx)

	// Set up test arguments
	args := AddArgs{
		Context:    ctx,
		DBConn:     conn,
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

	getArgs := GetArgs{
		Context:   ctx,
		DBConn:    conn,
		IndexPath: tempDir,
		Name:      testFilePath,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Get(getArgs)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Get command prints out the content and a new line in the end
	testContent = testContent + "\n"

	// Validate the output
	if output != testContent {
		t.Errorf("Expected output '%s', but got '%s'", testContent, output)
	}
}

func TestGet_Webpage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testWebpageURL := "http://echo.jsontest.com/title/lorem/content/ipsum"

	// Setup database connection
	ctx, conn, err := setupDatabaseForTesting()
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL database: %v", err)
	}
	defer closeTestingDatabase()
	defer conn.Close(ctx)

	// Set up test arguments
	args := AddArgs{
		Context:    ctx,
		DBConn:     conn,
		IndexPath:  tempDir,
		DataType:   "text",
		SourceType: "webpage",
		DataURIs:   []string{testWebpageURL},
	}

	// Call the Add function
	Add(args)

	// Check if the index file was created
	indexFilePath := path.Join(tempDir, indexFile)
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

	getArgs := GetArgs{
		Context:   ctx,
		DBConn:    conn,
		IndexPath: tempDir,
		Name:      testWebpageURL,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Get(getArgs)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Fetching webpage content
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

	// Validating the contents of the copy file
	if !strings.Contains(output, webpageContentStr) {
		t.Errorf("Failed to validate webpage copy: Expected \"%s\", got \"%s\"", webpageContent, output)
	}
}
