package subcommands
import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

const addedCopiesSubDir = "add_cache";
const addedFile = "added.list"

type AddedListEntry struct {
	Name string
	URI string
	DataType string
	SourceType string
	TimeFirstAdded string
}

type AddCacheEntry struct {
	AddedListEntry
	TimeLastRefreshed string
	Contents []byte
}

type AddArgs struct {
	IndexPath string
	DataType string
	SourceType string
	MakeCopy bool
	DataURIs []string
}

func Add(a AddArgs){
	fmt.Println(fmt.Sprintf("Add is not fully implemented. dataType: %s, dataURIs: %v", a.DataType, a.DataURIs))
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

			ale := AddedListEntry{
				Name: u,
				URI: u,
				DataType: a.DataType,
				SourceType: a.SourceType,
				TimeFirstAdded: time.Now().String(),
			}
			if alreadyAdded(ale, addedFile) {
				fmt.Println(fmt.Sprintf("File %s has already been added. Skipping without refresh.", u))
				continue
			}
			if a.MakeCopy {
				// We'd first do something type-related if we supported anything besides text.
				err = copyFile(u, path.Join(a.IndexPath, addedCopiesSubDir, ale.Name), ale)
				if err != nil {
					fmt.Println(fmt.Sprintf("File %s failed to copy with err: %s. Skipping.", u, err))
					continue
				}
			}
			err = commitAdd(ale, addedFile)
			if err != nil {
				fmt.Println(fmt.Sprintf("File %s failed to commit with err: %s. Skipping.", u, err))
				continue
			}
		}
	default:
		fmt.Println(fmt.Sprintf("Invalid 'add' SourceType subsubcommand: %s", a.SourceType))
		os.Exit(1)
	}
}

func alreadyAdded(ale AddedListEntry, fp *os.File) bool {
	// Seek to the beginning of the file
	_, err := fp.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Printf("Failed to seek to beginning of file: %s\n", err)
		os.Exit(1)
	}

	decoder := gob.NewDecoder(fp)
	for {
		var entry AddedListEntry
		err := decoder.Decode(&entry)
		if err == io.EOF {
			// Reached end of file, entry not found
			break
		}
		if err != nil {
			fmt.Printf("Failed to decode entry: %s\n", err)
			os.Exit(1)
		}
		if entry.Name == ale.Name {
			return true
		}
	}
	return false
}

func commitAdd(ale AddedListEntry, fp *os.File) error {
	// Seek to the end of the file
	_, err := fp.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	// Create an encoder that will append to the file
	encoder := gob.NewEncoder(fp)

	// Encode and append the new entry
	err = encoder.Encode(ale)
	if err != nil {
		return fmt.Errorf("failed to encode and append new entry: %w", err)
	}

	return nil
}

func copyFile(src string, dest string, ale AddedListEntry) error {
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

	// Create the AddCacheEntry
	entry := AddCacheEntry{
		AddedListEntry: ale,
		TimeLastRefreshed: time.Now().UTC().Format(time.RFC3339),
		Contents:          content,
	}

	// Create destination dir
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0770); err != nil {
        return fmt.Errorf("failed to create destination dir $s: %w", dir, err)
    }
	// Create or open the destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Create a gob encoder
	encoder := gob.NewEncoder(destFile)

	// Encode and write the AddCacheEntry
	err = encoder.Encode(entry)
	if err != nil {
		return fmt.Errorf("failed to encode and write entry: %w", err)
	}

	return nil
}