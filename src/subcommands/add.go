package subcommands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	pb "accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const addedCopiesSubDir = "add_cache"
const addedFile = "added.list"

type AddedListEntry struct {
	Name           string
	URI            string
	DataType       string
	SourceType     string
	TimeFirstAdded string
}

type AddCacheEntry struct {
	AddedListEntry
	TimeLastRefreshed string
	Contents          []byte
}

type AddArgs struct {
	IndexPath  string
	DataType   string
	SourceType string
	MakeCopy   bool
	DataURIs   []string
}

func Add(a AddArgs) {
	// fmt.Println(fmt.Sprintf("Add is not fully implemented. dataType: %s, dataURIs: %v", a.DataType, a.DataURIs))
	switch a.SourceType {
	case "file":
		addedFilePath := path.Join(a.IndexPath, addedFile)
		addedFile, err := os.OpenFile(addedFilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to create or open file tracking added data at %s: %s", addedFilePath, err))
			os.Exit(1)
		}
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

			if alreadyAdded(u, addedFilePath) {
				fmt.Println(fmt.Sprintf("File %s has already been added. Skipping without refresh.", u))
				continue
			}

			ale := &pb.IndexListEntry{
				Name:             u,
				URI:              u,
				dataType:         a.DataType,
				sourceType:       a.SourceType,
				first_added_time: timestamppb.Now(),
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

func alreadyAdded(name string, filepath string) bool {
	// Seek to the beginning of the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Errorf("failed to read index file: %w", err)
		return false
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		fmt.Errorf("failed to marshall index file: %w", err)
	}

	for _, entry := range index.Entries {
		if entry.Name == name {
			return true
		}
	}

	return false
}

func commitAdd(ale pb.IndexListEntry, filepath string) error {

	var index pb.Index
	data, err := os.ReadFile(filepath)

	if err == nil {
		if err := proto.Unmarshal(data, &index); err != nil {
			return fmt.Errorf("failed to unmarshal existing index: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	index.Entries = append(index.Entries, ale)

	updatedData, err := proto.Marshal(&index)
	if err != nil {
		return fmt.Errorf("failed to marshal updated index: %w", err)
	}

	if err := os.WriteFile(filepath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated index to file: %w", err)
	}

	return nil
}

func copyFile(src string, dest string, ale pb.IndexListEntry) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Read the entire file content
	content, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Update the IndexListEntry
	ale.Content = content
	ale.LastRefreshedTime = timestamppb.Now()

	// Create destination dir
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("failed to create destination dir $s: %w", dir, err)
	}

	// Serialize the IndexListEntry
	data, err := proto.Marshal(&ale)
	if err != nil {
		return fmt.Errorf("failed to marshal IndexListEntry: %w", err)
	}

	// Write it to the destination
	//TODO: check the path properly
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}
