package tests

import (
	"bytes"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"accretional.com/semantifly/subcommands"
)

func TestGetCommand(t *testing.T) {
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

	// Create a mock `AddArgs` structure and add the file
	addArgs := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   "text",
		SourceType: "file",
		MakeCopy:   true,
		DataURIs:   []string{srcFile.Name()},
	}

	// Invoking the `Add` function
	subcommands.Add(addArgs)

	// Create a mock `GetArgs` structure
	getArgs := subcommands.GetArgs{
		IndexPath:  indexDir,
		DataType:   "text",
		SourceType: "file",
		Name:       filepath.Base(srcFile.Name()),
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	subcommands.Get(getArgs)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	if output != string(originalContent) {
		t.Errorf("Expected output '%s', but got '%s'", string(originalContent), output)
	}

	verifyTimeLastRefreshed(t, filepath.Join(cacheDir, filepath.Base(srcFile.Name())))
}

func verifyTimeLastRefreshed(t *testing.T, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	var entry subcommands.AddCacheEntry
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&entry)
	if err != nil {
		t.Fatalf("Failed to decode file %s: %v", filePath, err)
	}

	if entry.TimeLastRefreshed == entry.TimeFirstAdded {
		t.Errorf("TimeLastRefreshed was not updated")
	}
}

func TestGetNonExistentFile(t *testing.T) {
	indexDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(indexDir)

	getArgs := subcommands.GetArgs{
		IndexPath:  indexDir,
		DataType:   "text",
		SourceType: "file",
		Name:       "non_existent_file.txt",
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	subcommands.Get(getArgs)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	expectedOutput := "File 'non_existent_file.txt' not found in the index."
	if output != expectedOutput {
		t.Errorf("Expected output '%s', but got '%s'", expectedOutput, output)
	}
}