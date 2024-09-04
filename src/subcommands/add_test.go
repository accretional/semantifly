package subcommands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	database "accretional.com/semantifly/database"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func setupPostgres() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory: %v", err)
	}

	if filepath.Base(currentDir) != "semantifly" {
		err = os.Chdir("../..")
		if err != nil {
			return fmt.Errorf("Failed to change directory: %v", err)
		}
	}

	cmd := exec.Command("bash", "setup_postgres.sh")

	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to setup PostgreSQL server: %v", err)
	}

	return nil
}

func createTestingDatabase() (*pg.DB, error) {
	// Connect to the default "postgres" database
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Addr:     "localhost:5432",
		Database: "postgres",
	})

	// Drop the database if it exists, then create it
	_, err := db.Exec("DROP DATABASE IF EXISTS testdb")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to drop existing test database: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE testdb")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create test database: %v", err)
	}

	// Close the connection to the "postgres" database
	db.Close()

	// Connect to the newly created database
	testDB := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Addr:     "localhost:5432",
		Database: "testdb",
	})

	return testDB, nil
}

func closeTestingDatabase() error {
	// Connect to the default "postgres" database to drop the test database
	defaultDB := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "postgres",
		Addr:     "localhost:5432",
		Database: "postgres",
	})
	defer defaultDB.Close()

	// Terminate all connections to the test database
	_, err := defaultDB.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'testdb'")
	if err != nil {
		return fmt.Errorf("Failed to terminate connections to test database: %v", err)
	}

	// Drop the test database
	_, err = defaultDB.Exec("DROP DATABASE IF EXISTS testdb")
	if err != nil {
		return fmt.Errorf("Failed to drop test database: %v", err)
	}

	return nil
}

func setupDatabaseForTesting() (context.Context, database.PgxIface, error) {
	err := setupPostgres()

	if err != nil {
		return nil, nil, fmt.Errorf("setupPostgres failed: %v", err)
	}

	// Set a mock DATABASE_URL for testing
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/testdb")
	defer os.Unsetenv("DATABASE_URL")

	db, err := createTestingDatabase()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create test database: %v", err)
	}
	defer db.Close()

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

	testFilePath2 := path.Join(tempDir, "test_file.txt")

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
		DataURIs:   []string{testFilePath1, testFilePath2},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Add(args)

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
		MakeCopy:   true,
		DataURIs:   []string{testWebpageURL},
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
