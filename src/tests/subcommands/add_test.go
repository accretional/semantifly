package tests

import (
	"bytes"
	"encoding/gob"
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

	// Preparing the test data
	originalContent := []byte("Test Copy File")
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

	// Verifying the destination file
	dstFilePath := filepath.Join(cacheDir, srcFile.Name())
	dstFile, err := os.Open(dstFilePath)
	if err != nil {
		t.Fatalf("Failed to open destination file: %v", err)
	}
	defer dstFile.Close()

	// Decoding
	var entry subcommands.AddCacheEntry
	decoder := gob.NewDecoder(dstFile)
	err = decoder.Decode(&entry)
	if err != nil {
		t.Fatalf("Failed to decode destination file: %v", err)
	}

	// Assertions
	ale := subcommands.AddedListEntry{
		Name:       srcFile.Name(),
		URI:        srcFile.Name(),
		DataType:   "text",
		SourceType: "file",
	}

	if entry.Name != ale.Name {
		t.Errorf("Expected Name %s, got %s", ale.Name, entry.Name)
	}
	if entry.URI != ale.URI {
		t.Errorf("Expected URI %s, got %s", ale.URI, entry.URI)
	}
	if entry.DataType != ale.DataType {
		t.Errorf("Expected DataType %s, got %s", ale.DataType, entry.DataType)
	}
	if entry.SourceType != ale.SourceType {
		t.Errorf("Expected SourceType %s, got %s", ale.SourceType, entry.SourceType)
	}
	if !bytes.Equal(entry.Contents, originalContent) {
		t.Errorf("Expected Contents %s, got %s", string(originalContent), string(entry.Contents))
	}
}
