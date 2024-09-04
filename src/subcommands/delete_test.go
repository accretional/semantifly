package subcommands

import (
	"fmt"
	"os"
	"path"
	"testing"
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

	// Call the Add function
	Add(args)

	// Test case
	testArgs := DeleteArgs{
		Context:    ctx,
		DBConn:     conn,
		IndexPath:  tempDir,
		DeleteCopy: true,
		DataURIs:   []string{testFilePath1},
	}

	// Run the Delete function
	Delete(testArgs)

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

func TestDeleteRows(t *testing.T) {

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

	// Call the Add function
	Add(args)

	// Test case
	testArgs := DeleteArgs{
		Context:    ctx,
		DBConn:     conn,
		IndexPath:  tempDir,
		DeleteCopy: true,
		DataURIs:   []string{testFilePath1},
	}

	// Run the Delete function
	Delete(testArgs)

	// Verify the results
	names := []string{testFilePath1, testFilePath2}
	rows, err := conn.Query(ctx, `SELECT name FROM index_list WHERE name=ANY($1)`, names)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	defer rows.Close()

	deleted := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		deleted[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Failed during rows iteration: %v", err)
	}

	// Check if testFilePath1 was deleted
	if deleted[testFilePath1] {
		t.Errorf("test_name1 was not deleted from the database")
	}

	// Check if testFilePath2 still exists
	if !deleted[testFilePath2] {
		t.Errorf("test_name2 was unexpectedly deleted from the database")
	}
}
