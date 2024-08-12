package subcommands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const addedCopiesSubDir = "add_cache"
const indexFile = "index.list"

type AddArgs struct {
	IndexPath  string
	DataType   pb.DataType
	SourceType pb.SourceType
	MakeCopy   bool
	DataURIs   []string
}

// Adds the specified files to the index. It checks if each file can be added, and if so, creates an IndexListEntry for the file and commits the addition to the index. If MakeCopy is true, it also copies the file to the added copies subdirectory.
// Parameters:
//
//	a AddArgs: The arguments for the add operation, including the source type, index path, data URIs, data type, and make copy flag.
//
// Exceptions/Errors:
//   - If the file does not exist, it prints a message and continues to the next file.
//   - If the file is a directory, it prints a message and continues to the next file.
//   - If the file is not a regular file, it prints a message and continues to the next file.
//   - If an error occurs while checking if the file has already been added, it prints an error message and continues to the next file.
//   - If the file has already been added, it prints a message and continues to the next file.
//   - If the file fails to copy when MakeCopy is true, it prints an error message and continues to the next file.
//   - If the file fails to commit, it prints an error message and continues to the next file.
func Add(a AddArgs) {
	switch a.SourceType {
	case pb.SourceType_LOCAL_FILE:
		// Construct the index file path
		indexFilePath := path.Join(a.IndexPath, indexFile)
		for i, u := range a.DataURIs {
			f, err := os.Stat(u)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Failed to add file %s at input list index %v: file does not exist\n", u, i)
				continue
			}
			if f.IsDir() {
				fmt.Printf("Cannot add directory %s as file. Try adding as a directory instead. Skipping.\n", u)
				continue
			}
			if !f.Mode().IsRegular() {
				fmt.Printf("File %s is not a regular file and cannot be added. Skipping.\n", u)
				continue
			}

			if added, err := alreadyAdded(u, indexFilePath); err != nil {
				fmt.Printf("Error checking if file is already added: %v\n", err)
				continue
			} else if added {
				fmt.Printf("File %s has already been added. Skipping without refresh.\n", u)
				continue
			}

			ile := &pb.IndexListEntry{
				Name:           u,
				URI:            u,
				DataType:       a.DataType,
				SourceType:     a.SourceType,
				FirstAddedTime: timestamppb.Now(),
			}

			if a.MakeCopy {
				err = copyFile(u, path.Join(a.IndexPath, addedCopiesSubDir, ile.Name), ile)
				if err != nil {
					fmt.Printf("File %s failed to copy with err: %s. Skipping.\n", u, err)
					continue
				}
			}
			err = commitAdd(ile, indexFilePath)
			if err != nil {
				fmt.Printf("File %s failed to commit with err: %s. Skipping.\n", u, err)
				continue
			}
			fmt.Printf("Added file successfully: %s\n", u)
		}
	default:
		fmt.Printf("Invalid 'add' SourceType subsubcommand: %s\n", a.SourceType)
		os.Exit(1)
	}
}

// alreadyAdded checks if the given name is already present in the index file.
// Parameters:
// - name: the name to be checked in the index file
// - indexFilePath: the file path of the index file
// Returns:
// - bool: true if the name is already present in the index file, false otherwise
func alreadyAdded(name string, indexFilePath string) (bool, error) {
	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		// If the index file does not exist, return false
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read index file: %w", err)
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return false, fmt.Errorf("failed to marshall index file: %w", err)
	}

	for _, entry := range index.Entries {
		if entry.Name == name {
			return true, nil
		}
	}

	// If the name is not found in the index file, return false
	return false, nil
}

// commitAdd appends the provided IndexListEntry to the index file specified by indexFilePath.
// It reads the existing index file data, unmarshals it, appends the provided entry, marshals the updated index data,
// and writes it back to the index file. It returns an error if any operation fails.
//
// Parameters:
//
//	ile *pb.IndexListEntry: The IndexListEntry to be appended to the index entries.
//	indexFilePath string: The file path of the index file.
func commitAdd(ile *pb.IndexListEntry, indexFilePath string) error {

	var index pb.Index
	data, err := os.ReadFile(indexFilePath)

	if err == nil {
		if err := proto.Unmarshal(data, &index); err != nil {
			return fmt.Errorf("failed to unmarshal existing index: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	index.Entries = append(index.Entries, ile)

	updatedData, err := proto.Marshal(&index)
	if err != nil {
		return fmt.Errorf("failed to marshal updated index: %w", err)
	}
	if err := os.WriteFile(indexFilePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated index to file: %w", err)
	}

	return nil
}

// copyFile copies the content of the source file to the destination file and updates the given IndexListEntry with the content and last refreshed time.
// Parameters:
//
//	src: the path of the source file to be copied
//	dest: the path of the destination file where the content will be copied to
//	ile: a pointer to an IndexListEntry to be written to the destination file
func copyFile(src string, dest string, ile *pb.IndexListEntry) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	ile.Content = string(content)
	ile.LastRefreshedTime = timestamppb.Now()

	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("failed to create destination dir %s: %w", dir, err)
	}

	data, err := proto.Marshal(ile)
	if err != nil {
		return fmt.Errorf("failed to marshal IndexListEntry: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}
