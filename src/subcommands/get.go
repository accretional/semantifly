package subcommands

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

type GetArgs struct {
	IndexPath  string
	DataType   string
	SourceType string
	Name       string
}

func Get(g GetArgs) {
	addedFilePath := path.Join(g.IndexPath, addedFile)
	addedFile, err := os.OpenFile(addedFilePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file tracking added data at %s: %s\n", addedFilePath, err)
		os.Exit(1)
	}
	defer addedFile.Close()

	entry, found := findEntry(g.Name, addedFile)
	if !found {
		fmt.Printf("File '%s' not found in the index.\n", g.Name)
		return
	}

	if entry.SourceType != g.SourceType || entry.DataType != g.DataType {
		fmt.Printf("Mismatch in SourceType or DataType. Found: %s, %s. Requested: %s, %s\n",
			entry.SourceType, entry.DataType, g.SourceType, g.DataType)
		return
	}

	switch g.SourceType {
	case "file":
		err = printFileContents(entry, g.IndexPath)
		if err != nil {
			fmt.Printf("Failed to retrieve and print file contents: %s\n", err)
			return
		}
	default:
		fmt.Printf("Invalid 'get' SourceType: %s\n", g.SourceType)
		os.Exit(1)
	}
}

func findEntry(name string, fp *os.File) (AddedListEntry, bool) {
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
			break
		}
		if err != nil {
			fmt.Printf("Failed to decode entry: %s\n", err)
			os.Exit(1)
		}
		if entry.Name == name || entry.URI == name {
			return entry, true
		}
	}
	return AddedListEntry{}, false
}

func printFileContents(entry AddedListEntry, indexPath string) error {
	cachePath := path.Join(indexPath, addedCopiesSubDir, entry.Name)
	cacheFile, err := os.Open(cachePath)
	if err != nil {
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer cacheFile.Close()

	decoder := gob.NewDecoder(cacheFile)
	var cacheEntry AddCacheEntry
	err = decoder.Decode(&cacheEntry)
	if err != nil {
		return fmt.Errorf("failed to decode cache entry: %w", err)
	}

	fmt.Println(string(cacheEntry.Contents))

	cacheEntry.TimeLastRefreshed = time.Now().UTC().Format(time.RFC3339)
	cacheFile.Close()

	cacheFile, err = os.Create(cachePath)
	if err != nil {
		return fmt.Errorf("failed to open cache file for updating: %w", err)
	}
	defer cacheFile.Close()

	encoder := gob.NewEncoder(cacheFile)
	err = encoder.Encode(&cacheEntry)
	if err != nil {
		return fmt.Errorf("failed to update cache entry: %w", err)
	}

	return nil
}