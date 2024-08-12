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

func verifyDeletedFileEntry(srcFileName string, addedFilePath string) error {

	addedFile, err := os.Open(addedFilePath)
	if err != nil {
		return err
	}
	defer addedFile.Close()

	data, err := os.ReadFile(addedFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Added list file %s missing", addedFilePath)
		}
		return fmt.Errorf("Failed to read index file: %w", err)
	}

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
	if entryFound {
		return fmt.Errorf("Entry %s found in added list. Not deleted\n", srcFileName)
	}

	return nil
}

func verifyDeleteCopy(dstFilePath string) error {

	_, err := os.ReadFile(dstFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Error in checking for copy file %s: %v.", dstFilePath, err)
		}
	} else {
		return fmt.Errorf("Error in deleting copy file %s: Copy file found.", dstFilePath)
	}

	return nil
}

func TestDelete(t *testing.T) {

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

	// Need to add the file first
	addArgs := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   false,
		DataURIs:   []string{srcFile.Name()},
	}

	subcommands.Add(addArgs)

	// Now delete the entry
	deleteArgs := subcommands.DeleteArgs{
		IndexPath:  indexDir,
		DeleteCopy: true,
		DataURIs:   []string{srcFile.Name()},
	}

	subcommands.Delete(deleteArgs)

	indexFilePath := filepath.Join(indexDir, indexFile)
	if err := verifyDeletedFileEntry(srcFile.Name(), indexFilePath); err != nil {
		t.Fatalf("Failed to verify delete command in index list: %v", err)
	}

	dstFilePath := filepath.Join(cacheDir, srcFile.Name())
	if err := verifyDeleteCopy(dstFilePath); err != nil {
		t.Fatalf("Failed to verify delete copy: %v", err)
	}
}
