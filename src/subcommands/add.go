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
const addedFile = "added.list"

type AddArgs struct {
	IndexPath  string
	DataType   pb.DataType
	SourceType pb.SourceType
	MakeCopy   bool
	DataURIs   []string
}

func Add(a AddArgs) {
	// fmt.Println(fmt.Sprintf("Add is not fully implemented. dataType: %s, dataURIs: %v", a.DataType, a.DataURIs))
	switch a.SourceType {
	case pb.SourceType_LOCAL_FILE:
		addedFilePath := path.Join(a.IndexPath, addedFile)
		for i, u := range a.DataURIs {
			f, err := os.Stat(u)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println(fmt.Sprintf("Failed to add file %s at input list index %v: file does not exist", u, i))
				return
			}
			if f.IsDir() {
				fmt.Println(fmt.Sprintf("Cannot add directory %s as file. Try adding as a directory instead. Skipping.", u))
				continue
			}
			if !f.Mode().IsRegular() {
				fmt.Println(fmt.Sprintf("File %s is not a regular file and cannot be added. Skipping.", u))
				continue
			}

			if added, err := alreadyAdded(u, addedFilePath); err != nil {
				fmt.Printf("Error checking if file is already added: %v\n", err)
				continue
			} else if added {
				fmt.Printf("File %s has already been added. Skipping without refresh.\n", u)
				continue
			}

			ale := &pb.IndexListEntry{
				Name:           u,
				URI:            u,
				DataType:       a.DataType,
				SourceType:     a.SourceType,
				FirstAddedTime: timestamppb.Now(),
			}

			if a.MakeCopy {
				// We'd first do something type-related if we supported anything besides text.
				err = copyFile(u, path.Join(a.IndexPath, addedCopiesSubDir, ale.Name), ale)
				if err != nil {
					fmt.Println(fmt.Sprintf("File %s failed to copy with err: %s. Skipping.", u, err))
					continue
				}
			}
			err = commitAdd(ale, addedFilePath)
			if err != nil {
				fmt.Println(fmt.Sprintf("File %s failed to commit with err: %s. Skipping.", u, err))
				continue
			}
			fmt.Println(fmt.Sprintf("Added file successfully: %s", u))
		}
	default:
		fmt.Println(fmt.Sprintf("Invalid 'add' SourceType subsubcommand: %s", a.SourceType))
		os.Exit(1)
	}
}

func alreadyAdded(name string, filepath string) (bool, error) {
	// Seek to the beginning of the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("Failed to read index file: %w", err)
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return false, fmt.Errorf("Failed to marshall index file: %w", err)
	}

	for _, entry := range index.Entries {
		if entry.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func commitAdd(ale *pb.IndexListEntry, filepath string) error {

	var index pb.Index
	data, err := os.ReadFile(filepath)

	if err == nil {
		if err := proto.Unmarshal(data, &index); err != nil {
			return fmt.Errorf("Failed to unmarshal existing index: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to read index file: %w", err)
	}

	index.Entries = append(index.Entries, ale)

	updatedData, err := proto.Marshal(&index)
	if err != nil {
		return fmt.Errorf("Failed to marshal updated index: %w", err)
	}

	if err := os.WriteFile(filepath, updatedData, 0644); err != nil {
		return fmt.Errorf("Failed to write updated index to file: %w", err)
	}

	return nil
}

func copyFile(src string, dest string, ale *pb.IndexListEntry) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Read the entire file content
	content, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("Failed to read source file: %w", err)
	}

	// Update the IndexListEntry
	ale.Content = string(content)
	ale.LastRefreshedTime = timestamppb.Now()

	// Create destination dir
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("Failed to create destination dir %s: %w", dir, err)
	}

	// Serialize the IndexListEntry
	data, err := proto.Marshal(ale)
	if err != nil {
		return fmt.Errorf("Failed to marshal IndexListEntry: %w", err)
	}

	// Write it to the destination
	//TODO: check the path properly
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("Failed to write to destination file: %w", err)
	}

	return nil
}
