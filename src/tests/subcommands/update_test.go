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

func verifyUpdatedFileEntry(addedFilePath string, ile *pb.IndexListEntry) error {

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
		if entry.Name == ile.Name {
			entryFound = true

			if entry.URI != ile.URI {
				return fmt.Errorf("Error in updating the entry: expected URI %s, got %s.\n", ile.URI, entry.URI)
			}
			if entry.DataType != ile.DataType {
				return fmt.Errorf("Error in updating the entry: expected DataType %s, got %s.\n", ile.DataType, entry.DataType)
			}
			if entry.SourceType != ile.SourceType {
				return fmt.Errorf("Error in updating the entry: expected SourceType %s, got %s.\n", ile.SourceType, entry.SourceType)
			}
		}
	}
	if !entryFound {
		return fmt.Errorf("Entry %s not found in index list.\n", ile.Name)
	}

	return nil
}

func verifyUpdateCopy(dstFilePath string, ile *pb.IndexListEntry) error {

	dest, err := os.ReadFile(dstFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s missing", dstFilePath)
		}
		return fmt.Errorf("failed to read index file: %w", err)
	}

	var entry pb.IndexListEntry
	if err := proto.Unmarshal(dest, &entry); err != nil {
		return fmt.Errorf("failed to unmarshall entry file: %w", err)
	}

	// Assertions
	if entry.Name != ile.Name {
		return fmt.Errorf("Error in %s: Expected Name %s, got %s", dstFilePath, ile.Name, entry.Name)
	}
	if entry.URI != ile.URI {
		return fmt.Errorf("Error in %s: Expected URI %s, got %s", dstFilePath, ile.URI, entry.URI)
	}
	if entry.DataType != ile.DataType {
		return fmt.Errorf("Error in %s: Expected DataType %s, got %s", dstFilePath, ile.DataType, entry.DataType)
	}
	if entry.SourceType != ile.SourceType {
		return fmt.Errorf("Error in %s: Expected SourceType %s, got %s", dstFilePath, ile.SourceType, entry.SourceType)
	}
	if entry.Content != ile.Content {
		return fmt.Errorf("Error in %s: Expected Content %s, got %s", dstFilePath, ile.Content, string(entry.Content))
	}

	return nil
}

func TestUpdate(t *testing.T) {

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

	// Create an updated source file
	updatedContent := "Test File Contents - Updated"
	updatedFile := createTempFile(t, indexDir, updatedContent)

	// Update the entry
	updateArgs := subcommands.UpdateArgs{
		IndexPath:  indexDir,
		Name:       srcFile.Name(),
		DataURI:    updatedFile.Name(),
		UpdateCopy: "true",
	}

	subcommands.Update(updateArgs)

	ile := &pb.IndexListEntry{
		Name:       srcFile.Name(),
		URI:        updatedFile.Name(),
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		Content:    updatedContent,
	}

	indexFilePath := filepath.Join(indexDir, indexFile)
	if err := verifyUpdatedFileEntry(indexFilePath, ile); err != nil {
		t.Fatalf("Failed to verify delete command in index list: %v", err)
	}

	dstFilePath := filepath.Join(cacheDir, srcFile.Name())
	if err := verifyUpdateCopy(dstFilePath, ile); err != nil {
		t.Fatalf("Failed to verify updated copy: %v", err)
	}
}
