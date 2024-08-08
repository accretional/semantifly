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

// createTempFile creates a temporary file in the specified directory with the given data.
// It returns a pointer to the created file. If an error occurs during file creation or writing,
// it fails the test with a fatal error message.
//
// Parameters:
//
//	t *testing.T: The testing.T instance for reporting test failures
//	dir string: The directory in which to create the temporary file
//	data string: The data to be written to the temporary file
//
// Return:
//
//	*os.File: A pointer to the created temporary file
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

// verifyAddedFileEntry checks if the given source file name exists in the added file list.
// It reads the added file list, unmarshals it into a protobuf Index structure,
// and searches for the source file name in the entries.
//
// Parameters:
//   - srcFileName: The name of the source file to search for in the added list.
//   - addedFilePath: The path to the added file list.
//
// Returns:
//   - error: nil if the entry is found, otherwise an error describing the issue.
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

// verifyMakeCopy verifies the content of the destination file against the provided IndexListEntry and content.
// It reads the destination file, unmarshalls its content into an IndexListEntry, and compares its fields
// with the provided IndexListEntry and content.
//
// Parameters:
// - dstFilePath: the file path of the destination file to be verified
// - ale: the IndexListEntry to compare with the destination file's content
// - content: the content to compare with the destination file's content field
//
// Returns:
// - error: an error indicating any issues encountered during the verification process
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

// TestReadWrite is a test function to verify the behavior of the ReadWrite functionality. It sets up the necessary paths and test data, invokes the `Add` function, and verifies the expected behavior by checking the added file entry and the contents of the copied file.
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

	const indexFile = "index.list"

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

	// Verifying the source file entry in index list
	indexFilePath := filepath.Join(indexDir, indexFile)

	if err := verifyAddedFileEntry(srcFile.Name(), indexFilePath); err != nil {
		t.Fatalf("Failed to verify source file entry in added list: %v", err)
	}

	// Creating dstFilePath and ale for assertions
	dstFilePath := filepath.Join(cacheDir, srcFile.Name())
	ile := &pb.IndexListEntry{
		Name:       srcFile.Name(),
		URI:        srcFile.Name(),
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
	}

	// Verifying the contents of the copy of source file
	if err := verifyMakeCopy(dstFilePath, ile, originalContent); err != nil {
		t.Fatalf("Failed to verify copy of source file: %v", err)
	}
}

// TestMultipleAddCommands tests the functionality of adding multiple source files to the index and verifying their entries and contents.
// It sets up temporary directories and files, creates mock `AddArgs` structures, invokes the `Add` function for each source file, verifies the entries in the added list, and checks the contents of the copied files in the cache directory.
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

	const indexFile = "index.list"

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

	// Creating the index file path
	indexFilePath := filepath.Join(indexDir, indexFile)

	// Verifying the source file entries in added list
	if err := verifyAddedFileEntry(srcFile1.Name(), indexFilePath); err != nil {
		t.Fatalf("Failed to verify source file 1 entry in added list: %v", err)
	}
	if err := verifyAddedFileEntry(srcFile2.Name(), indexFilePath); err != nil {
		t.Fatalf("Failed to verify source file 2 entry in added list: %v", err)
	}

	// Creating dstFilePath1 and ale for assertions of source file 1
	dstFilePath1 := filepath.Join(cacheDir, srcFile1.Name())
	ile := &pb.IndexListEntry{
		Name:       srcFile1.Name(),
		URI:        srcFile1.Name(),
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
	}

	// Verifying the contents of the copy of source file 1
	if err := verifyMakeCopy(dstFilePath1, ile, srcContent1); err != nil {
		t.Fatalf("Failed to verify copy of source file 1: %v", err)
	}

	// Creating dstFilePath2 and ale for assertions of source file 2
	dstFilePath2 := filepath.Join(cacheDir, srcFile2.Name())
	ile.Name = srcFile2.Name()
	ile.URI = srcFile2.Name()

	// Verifying the contents of the copy of source file 2
	if err := verifyMakeCopy(dstFilePath2, ile, srcContent2); err != nil {
		t.Fatalf("Failed to verify copy of source file 2: %v", err)
	}
}
