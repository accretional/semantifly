package subcommands

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const addedCopiesSubDir = "add_cache"
const indexFile = "index.list"

// readIndex reads an index file from the given path and returns a map of IndexListEntry objects.
// If the file is not found and ignoreIfNotFound is true, it returns an empty map.
// If the file is empty and ignoreIfNotFound is false, it returns an error.
//
// Parameters:
//   - indexFilePath: The path to the index file.
//   - ignoreIfNotFound: If true, ignore errors when the file is not found or empty.
//
// Returns:
//   - A map of string keys to IndexListEntry pointers. The name of the IndexListEntries are used as keys.
func readIndex(indexFilePath string, ignoreIfNotFound bool) (map[string]*pb.IndexListEntry, error) {
	index := &pb.Index{}
	data, err := os.ReadFile(indexFilePath)

	indexMap := make(map[string]*pb.IndexListEntry)

	if err != nil {
		if os.IsNotExist(err) && ignoreIfNotFound {
			return indexMap, nil
		}
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	if err := proto.Unmarshal(data, index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index file: %w", err)
	}

	if index == nil || len(index.Entries) == 0 {
		if !ignoreIfNotFound {
			return nil, fmt.Errorf("empty index file")
		} else {
			return indexMap, nil
		}
	}

	for _, entry := range index.Entries {
		indexMap[entry.Name] = entry
	}

	return indexMap, nil
}

// writeIndex writes the provided index map to the specified file path.
// It marshals the index entries into a protobuf format and saves it to the file.
//
// Parameters:
//   - indexFilePath: The path where the index file will be written.
//   - indexMap: A map containing index entries to be written.
func writeIndex(indexFilePath string, indexMap map[string]*pb.IndexListEntry) error {

	index := &pb.Index{
		Entries: make([]*pb.IndexListEntry, 0, len(indexMap)),
	}

	for _, entry := range indexMap {
		index.Entries = append(index.Entries, entry)
	}

	updatedData, err := proto.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal updated index: %w", err)
	}

	if err := os.WriteFile(indexFilePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated index to file: %w", err)
	}

	return nil
}

// makeCopy creates a copy of the given IndexListEntry in the specified index path.
// It updates the LastRefreshedTime of the entry and marshals it to a file.
//
// Parameters:
//   - indexPath: The base path where the copy will be stored.
//   - ile: Pointer to the IndexListEntry to be copied.
func makeCopy(indexPath string, ile *pb.IndexListEntry) error {
	fmt.Println("ile copy")
	fmt.Println(ile)

	content, err := fetch.FetchFromSource(ile.SourceType, ile.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch the content of URI %s: %v", ile.URI, err)

	}

	ileCopy := &pb.IndexListEntry{
		Name:              ile.Name,
		URI:               ile.URI,
		DataType:          ile.DataType,
		SourceType:        ile.SourceType,
		FirstAddedTime:    ile.FirstAddedTime,
		Content:           string(content),
		LastRefreshedTime: timestamppb.Now(),
	}

	dest := path.Join(indexPath, addedCopiesSubDir, ile.Name)
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("failed to create destination dir %s: %w", dir, err)
	}

	data, err := proto.Marshal(ileCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal IndexListEntry: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}

// fetchFromCopy retrieves the content of a file from the copy directory.
//
// Parameters:
//   - indexPath: The base path of the index.
//   - name: The name of the file to fetch.
//
// Returns:
//   - []byte: The content of the file as a byte slice.
func fetchFromCopy(indexPath string, name string) ([]byte, error) {
	copyFilePath := path.Join(indexPath, addedCopiesSubDir, name)

	_, err := os.Stat(copyFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file does not exist")
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read the file contents
	data, err := os.ReadFile(copyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", copyFilePath, err)
	}

	// Unmarshal the data into an IndexListEntry
	ile := &pb.IndexListEntry{}
	err = proto.Unmarshal(data, ile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal IndexListEntry: %w", err)
	}

	if ile.Content == "" {
		return nil, nil
	}

	return []byte(ile.Content), nil
}

// parseDataType converts a string representation of a data type to its corresponding pb.DataType enum value.
// It returns the parsed pb.DataType and an error if the input string is not a valid data type.
//
// Parameters:
//   - str: A string representing the data type to be parsed.
func parseDataType(str string) (pb.DataType, error) {
	val, ok := pb.DataType_value[strings.ToUpper(str)]
	if !ok {
		return pb.DataType_TEXT, fmt.Errorf("unknown data type: %s", str)
	}
	return pb.DataType(val), nil
}

// parseSourceType converts a string representation of a source type to its corresponding pb.SourceType enum value.
// It returns the parsed pb.SourceType and nil error if successful, or pb.SourceType_LOCAL_FILE and an error if the input is invalid.
//
// Parameters:
//   - str: A string representing the source type to be parsed.
func parseSourceType(str string) (pb.SourceType, error) {
	val, ok := pb.SourceType_value[strings.ToUpper(str)]
	if !ok {
		return pb.SourceType_LOCAL_FILE, fmt.Errorf("unknown source type: %s", str)
	}
	return pb.SourceType(val), nil
}
