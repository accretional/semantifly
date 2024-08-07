package tests

import (
	"bytes"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"testing"

	"accretional.com/semantifly/subcommands"
)

func createTempFile(t *testing.T, dir string, data []byte) *os.File {
	file, err := os.CreateTemp(dir, "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	return file
}

func verifyFileEntry(t *testing.T, dstFilePath string, srcFileName string, addedFilePath string, originalContent []byte) error {

	// Creating ale for assertions
	ale := subcommands.AddedListEntry{
		Name:       srcFileName,
		URI:        srcFileName,
		DataType:   "text",
		SourceType: "file",
	}

	//Verify file present in the Index
	addedFile, err := os.Open(addedFilePath)
	if err != nil {
		return err
	}
	defer addedFile.Close()

	_, err = addedFile.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Failed to decode entry: %v\n", err)
	}

	decoder := gob.NewDecoder(addedFile)
	entryFound := false

	for {
		var indexEntry subcommands.AddCacheEntry
		err = decoder.Decode(&indexEntry)
		if err == io.EOF {
			// Reached end of file, entry not found
			break
		}
		if err != nil {
			t.Fatalf("Failed to decode entry: %v\n", err)
			os.Exit(1)
		}
		if indexEntry.Name == ale.Name {
			entryFound = true
			break
		}
	}

	if !entryFound {
		t.Errorf("Entry %s not found in added list\n", srcFileName)
	}

	// Check destination file opening
	dstFile, err := os.Open(dstFilePath)
	if err != nil {
		t.Fatalf("Failed to open destination file %s: %v", dstFilePath, err)
	}
	defer dstFile.Close()

	// Decoding
	var entry subcommands.AddCacheEntry
	decoder = gob.NewDecoder(dstFile)
	err = decoder.Decode(&entry)
	if err != nil {
		t.Fatalf("Failed to decode destination file %s: %v", dstFilePath, err)
	}

	// Assertions
	if entry.Name != ale.Name {
		t.Errorf("Error in %s: Expected Name %s, got %s", dstFilePath, ale.Name, entry.Name)
	}
	if entry.URI != ale.URI {
		t.Errorf("Error in %s: Expected URI %s, got %s", dstFilePath, ale.URI, entry.URI)
	}
	if entry.DataType != ale.DataType {
		t.Errorf("Error in %s: Expected DataType %s, got %s", dstFilePath, ale.DataType, entry.DataType)
	}
	if entry.SourceType != ale.SourceType {
		t.Errorf("Error in %s: Expected SourceType %s, got %s", dstFilePath, ale.SourceType, entry.SourceType)
	}
	if !bytes.Equal(entry.Contents, originalContent) {
		t.Errorf("Error in %s: Expected Contents %s, got %s", dstFilePath, string(originalContent), string(entry.Contents))
	}

	return nil
}

func TestReadWrite(t *testing.T) {

	// Setting up the paths
	indexDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(indexDir)

	cacheDir := filepath.Join(indexDir, "add_cache")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	const addedFile = "added.list"

	// Preparing the test data
	originalContent := []byte("Test File Contents")
	srcFile := createTempFile(t, indexDir, originalContent)

	// Create a mock `AddArgs` structure
	args := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   "text",
		SourceType: "file",
		MakeCopy:   true,
		DataURIs:   []string{srcFile.Name()},
	}

	// Invoking the `Add` function
	subcommands.Add(args)

	// Verifying the file entry
	dstFilePath := filepath.Join(cacheDir, srcFile.Name())
	addedFilePath := filepath.Join(indexDir, addedFile)

	verifyFileEntry(t, dstFilePath, srcFile.Name(), addedFilePath, originalContent)
}

func TestMultipleAddCommands(t *testing.T) {
	// Setting up the paths
	indexDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(indexDir)

	cacheDir := filepath.Join(indexDir, "add_cache")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	srcContent1 := []byte("Test Content 1")
	srcContent2 := []byte("Test Content 2")

	srcFile1 := createTempFile(t, indexDir, srcContent1)
	srcFile2 := createTempFile(t, indexDir, srcContent2)

	const addedFile = "added.list"

	// Create a mock `AddArgs` structure to include source file 1
	args := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   "text",
		SourceType: "file",
		MakeCopy:   true,
		DataURIs:   []string{srcFile1.Name()},
	}

	// Invoking the `Add` function for source file 1
	subcommands.Add(args)

	// Modifying the args to now include source file 2
	args.DataURIs = []string{srcFile2.Name()}

	// Invoking the `Add` function for source file 2
	subcommands.Add(args)

	// Destination paths for the two files
	dstFilePath1 := filepath.Join(cacheDir, srcFile1.Name())
	dstFilePath2 := filepath.Join(cacheDir, srcFile2.Name())

	// Index file path
	addedFilePath := filepath.Join(indexDir, addedFile)

	// Verifying the file entries
	verifyFileEntry(t, dstFilePath1, srcFile1.Name(), addedFilePath, srcContent1)
	verifyFileEntry(t, dstFilePath2, srcFile2.Name(), addedFilePath, srcContent2)
}
