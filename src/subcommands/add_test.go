package subcommands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"accretional.com/semantifly/database"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func setupDatabaseForTesting() (context.Context, database.PgxIface, error) {

	// Test connection
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to establish connection to the database: %v", err)
	}

	// Test database table initialisation
	err = database.InitializeDatabaseSchema(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to initialise the database schema: %v", err)
	}

	return ctx, conn, nil
}

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

	// Setup testing database tables
	ctx, conn, err := setupDatabaseForTesting()
	if err != nil {
		t.Fatalf("Failed to setup the testing database: %v", err)
	}
	defer conn.Close(ctx)

	testFileData := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath,
	}

	args := &pb.AddRequest{
		AddedMetadata: testFileData,
		MakeCopy:      true,
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Call the Add function with the buffer
	err = SubcommandAdd(ctx, conn, args, tempDir, &buf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
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

	// Check if the test file was added to the index
	if entry, exists := indexMap[testFilePath]; !exists {
		t.Errorf("Test file was not added to the index")
	} else {
		// Verify the entry details
		if entry.Name != testFilePath {
			t.Errorf("Expected Name %s, got %s", testFilePath, entry.Name)
		}
		if entry.ContentMetadata.URI != testFilePath {
			t.Errorf("Expected URI %s, got %s", testFilePath, entry.ContentMetadata.URI)
		}
		if entry.ContentMetadata.DataType != pb.DataType_TEXT {
			t.Errorf("Expected DataType %v, got %v", pb.DataType_TEXT, entry.ContentMetadata.DataType)
		}
		if entry.ContentMetadata.SourceType != pb.SourceType_LOCAL_FILE {
			t.Errorf("Expected SourceType %v, got %v", pb.SourceType_LOCAL_FILE, entry.ContentMetadata.SourceType)
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

	// Setup testing database tables
	ctx, conn, err := setupDatabaseForTesting()
	if err != nil {
		t.Fatalf("Failed to setup the testing database: %v", err)
	}
	defer conn.Close(ctx)

	testFilePath2 := path.Join(tempDir, "test_file.txt")

	// Set up test arguments

	testFileData1 := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath1,
	}

	args := &pb.AddRequest{
		AddedMetadata: testFileData1,
		MakeCopy:      true,
	}

	var buf1 bytes.Buffer
	err = SubcommandAdd(ctx, conn, args, tempDir, &buf1)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	testFileData2 := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath2,
	}

	args = &pb.AddRequest{
		AddedMetadata: testFileData2,
		MakeCopy:      true,
	}

	var buf2 bytes.Buffer
	err = SubcommandAdd(ctx, conn, args, tempDir, &buf2)
	if err == nil {
		t.Fatalf("Add function did not return an error when it was suppposed to.")
	}
	output := buf2.String()

	// Checking if the second entry was skipped
	if !strings.Contains(err.Error(), "Skipping without refresh") {
		t.Errorf("Expected error 'Skipping without refresh', but got '%s'", output)
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

	// Setup testing database tables
	ctx, conn, err := setupDatabaseForTesting()
	if err != nil {
		t.Fatalf("Failed to setup the testing database: %v", err)
	}
	defer conn.Close(ctx)

	// Create the test url
	testWebpageURL := "http://echo.jsontest.com/title/lorem/content/ipsum"

	// Set up test arguments
	webData := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 1,
		URI:        testWebpageURL,
	}

	args := &pb.AddRequest{
		AddedMetadata: webData,
		MakeCopy:      true,
	}

	var buf bytes.Buffer

	err = SubcommandAdd(ctx, conn, args, tempDir, &buf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
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

	// Check if the test file was added to the index
	if entry, exists := indexMap[testWebpageURL]; !exists {
		t.Errorf("Test file was not added to the index")
	} else {
		// Verify the entry details
		if entry.Name != testWebpageURL {
			t.Errorf("Expected Name %s, got %s", testWebpageURL, entry.Name)
		}
		if entry.ContentMetadata.URI != testWebpageURL {
			t.Errorf("Expected URI %s, got %s", testWebpageURL, entry.ContentMetadata.URI)
		}
		if entry.ContentMetadata.DataType != pb.DataType_TEXT {
			t.Errorf("Expected DataType %v, got %v", pb.DataType_TEXT, entry.ContentMetadata.DataType)
		}
		if entry.ContentMetadata.SourceType != pb.SourceType_WEBPAGE {
			t.Errorf("Expected SourceType %v, got %v", pb.SourceType_LOCAL_FILE, entry.ContentMetadata.SourceType)
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
}

func TestAdd_Database(t *testing.T) {
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

	// Setup database connection
	ctx, conn, err := setupDatabaseForTesting()
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL database: %v", err)
	}
	defer conn.Close(ctx)

	testFileData := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        testFilePath,
	}

	args := &pb.AddRequest{
		AddedMetadata: testFileData,
		MakeCopy:      true,
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Call the Add function with the buffer
	err = SubcommandAdd(ctx, conn, args, tempDir, &buf)
	if err != nil {
		t.Fatalf("Add function returned an error: %v", err)
	}

	// Check if the test index was added to the database
	var jsonIndex []byte

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("unable to connect to database: %v", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
			SELECT entry
			FROM index_list 
			WHERE name=$1
		`, testFilePath).Scan(&jsonIndex)

	if err != nil {
		t.Fatalf("Failed to read the index from the database: %v", err)
	}

	var ile pb.IndexListEntry
	if err = protojson.Unmarshal(jsonIndex, &ile); err != nil {
		t.Fatalf("failed to unmarshal content metadata JSON to protobuf: %v", err)
	}

	if ile.Name != testFilePath {
		t.Errorf("Expected Name %s, got %s", testFilePath, ile.Name)
	}
	if ile.ContentMetadata.URI != testFilePath {
		t.Errorf("Expected URI %s, got %s", testFilePath, ile.ContentMetadata.URI)
	}
	if ile.ContentMetadata.DataType != pb.DataType_TEXT {
		t.Errorf("Expected DataType %v, got %v", pb.DataType_TEXT, ile.ContentMetadata.DataType)
	}
	if ile.ContentMetadata.SourceType != pb.SourceType_LOCAL_FILE {
		t.Errorf("Expected SourceType %v, got %v", pb.SourceType_LOCAL_FILE, ile.ContentMetadata.SourceType)
	}
	if ile.FirstAddedTime == nil {
		t.Errorf("FirstAddedTime is nil")
	}
}
