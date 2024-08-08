package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"accretional.com/semantifly/subcommands"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

func createTempFile(t *testing.T, dir string, data string) *os.File {
	file, err := os.CreateTemp(dir, "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()
	_, err = file.WriteString(data)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	return file
}

func verifyAddedFileEntry(srcFileName string, addedFilePath string) error {

	//Verify source list entry present in the Index
	addedFile, err := os.Open(addedFilePath)
	if err != nil {
		return err
	}
	defer addedFile.Close()

	// Reading the added.list file
	data, err := os.ReadFile(addedFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Added list file %s missing", addedFilePath)
		}
		return fmt.Errorf("Failed to read index file: %w", err)
	}

	// Checking for srcFileName in added.list file
	var entryFound = false
	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("Failed to marshall index file: %w", err)
	}

	for _, entry := range index.Entries {
		if entry.Name == srcFileName {
			entryFound = true
		}
	}
	if !entryFound {
		return fmt.Errorf("Entry %s not found in added list\n", srcFileName)
	}

	return nil
}

func verifyMakeCopy(dstFilePath string, ale *pb.IndexListEntry, content string) error {
	// Check destination file opening
	dest, err := os.ReadFile(dstFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s missing", dstFilePath)
		}
		return fmt.Errorf("failed to read index file: %w", err)
	}

	// Marshall entry file
	var entry pb.IndexListEntry
	if err := proto.Unmarshal(dest, &entry); err != nil {
		return fmt.Errorf("failed to unmarshall entry file: %w", err)
	}

	// Assertions
	if entry.Name != ale.Name {
		return fmt.Errorf("Error in %s: Expected Name %s, got %s", dstFilePath, ale.Name, entry.Name)
	}
	if entry.URI != ale.URI {
		return fmt.Errorf("Error in %s: Expected URI %s, got %s", dstFilePath, ale.URI, entry.URI)
	}
	if entry.DataType != ale.DataType {
		return fmt.Errorf("Error in %s: Expected DataType %s, got %s", dstFilePath, ale.DataType, entry.DataType)
	}
	if entry.SourceType != ale.SourceType {
		return fmt.Errorf("Error in %s: Expected SourceType %s, got %s", dstFilePath, ale.SourceType, entry.SourceType)
	}
	if entry.Content != content {
		return fmt.Errorf("Error in %s: Expected Contents %s, got %s", dstFilePath, content, string(entry.Content))
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
	originalContent := "Test File Contents"
	srcFile := createTempFile(t, indexDir, originalContent)

	// Create a mock `AddArgs` structure
	args := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   true,
		DataURIs:   []string{srcFile.Name()},
	}

	// Invoking the `Add` function
	subcommands.Add(args)

	// Verifying the source file entry in added list
	addedFilePath := filepath.Join(indexDir, addedFile)

	if err := verifyAddedFileEntry(srcFile.Name(), addedFilePath); err != nil {
		t.Fatalf("Failed to verify source file entry in added list: %v", err)
	}

	// Creating dstFilePath and ale for assertions
	dstFilePath := filepath.Join(cacheDir, srcFile.Name())
	ale := &pb.IndexListEntry{
		Name:       srcFile.Name(),
		URI:        srcFile.Name(),
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
	}

	// Verifying the contents of the copy of source file
	if err := verifyMakeCopy(dstFilePath, ale, originalContent); err != nil {
		t.Fatalf("Failed to verify copy of source file: %v", err)
	}
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

	// Setting up contents of two source files
	srcContent1 := "Test Content 1"
	srcContent2 := "Test Content 2"

	srcFile1 := createTempFile(t, indexDir, srcContent1)
	srcFile2 := createTempFile(t, indexDir, srcContent2)

	const addedFile = "added.list"

	// Create a mock `AddArgs` structure to include source file 1
	args := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   true,
		DataURIs:   []string{srcFile1.Name()},
	}

	// Invoking the `Add` function for source file 1
	subcommands.Add(args)

	// Modifying the args to now include source file 2
	args.DataURIs = []string{srcFile2.Name()}

	// Invoking the `Add` function for source file 2
	subcommands.Add(args)

	// Index file path
	addedFilePath := filepath.Join(indexDir, addedFile)

	// Verifying the source file entries in added list
	if err := verifyAddedFileEntry(srcFile1.Name(), addedFilePath); err != nil {
		t.Fatalf("Failed to verify source file 1 entry in added list: %v", err)
	}
	if err := verifyAddedFileEntry(srcFile2.Name(), addedFilePath); err != nil {
		t.Fatalf("Failed to verify source file 2 entry in added list: %v", err)
	}

	// Creating dstFilePath1 and ale for assertions of source file 1
	dstFilePath1 := filepath.Join(cacheDir, srcFile1.Name())
	ale := &pb.IndexListEntry{
		Name:       srcFile1.Name(),
		URI:        srcFile1.Name(),
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
	}

	// Verifying the contents of the copy of source file 1
	if err := verifyMakeCopy(dstFilePath1, ale, srcContent1); err != nil {
		t.Fatalf("Failed to verify copy of source file 1: %v", err)
	}

	// Creating dstFilePath2 and ale for assertions of source file 2
	dstFilePath2 := filepath.Join(cacheDir, srcFile2.Name())
	ale.Name = srcFile2.Name()
	ale.URI = srcFile2.Name()

	// Verifying the contents of the copy of source file 2
	if err := verifyMakeCopy(dstFilePath2, ale, srcContent2); err != nil {
		t.Fatalf("Failed to verify copy of source file 2: %v", err)
	}
}
