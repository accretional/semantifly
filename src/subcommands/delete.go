package subcommands

import (
	"fmt"
	"os"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

type DeleteArgs struct {
	IndexPath  string
	DeleteCopy bool
	DataURIs   []string
}

// Delete removes the specified data URIs from the index file at the given index path.
// It optionally deletes the associated data files as well based on the DeleteCopy flag.
//
// Parameters:
//   - d.IndexPath: The path to the directory containing the index file.
//   - d.DataURIs: A slice of data URIs to be deleted from the index.
//   - d.DeleteCopy: A boolean flag indicating whether to delete the associated data files.
//
// Errors:
//   - If there is an error searching for a URI in the index, an error message is printed and the URI is skipped.
//   - If there is an error deleting a URI from the index, an error message is printed and the URI is skipped.
//   - If there is an error deleting the associated data file, an error message is printed.
func Delete(d DeleteArgs) {
	indexFilePath := path.Join(d.IndexPath, indexFile)

	for _, uri := range d.DataURIs {
		if inIndex, err := isInIndex(uri, indexFilePath); err != nil {
			fmt.Printf("Failed to search for file %s in index %s with err: %s, skipping", uri, indexFilePath, err)
		} else if inIndex {
			if err := deleteFromIndex(uri, indexFilePath); err != nil {
				fmt.Printf("Failed to remove file %s from index %s with err: %s, skipping", uri, indexFilePath, err)
				continue
			} else {
				fmt.Printf("Deleted entry from index: %s\n", uri)
			}

			if d.DeleteCopy {
				if err := deleteCopy(d.IndexPath, uri); err != nil {
					fmt.Printf("Failed to delete copy of file %s with err: %s, skipping", uri, err)
				}
			}
		} else {
			fmt.Printf("Entry not found in index: %s, skipping\n", uri)
			continue
		}
	}

	fmt.Println("Delete operation completed.")
}

// isInIndex checks if a given name exists in the specified index file.
// It reads the index file, unmarshals the protobuf data, and searches for the name in the index entries.
//
// Parameters:
//   - name: The name to search for in the index.
//   - indexFilePath: The path to the index file.
//
// Returns:
//   - bool: True if the name is found in the index, false otherwise.
func isInIndex(name string, indexFilePath string) (bool, error) {
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
	return false, nil
}

// deleteFromIndex deletes an entry with the given name from the index file at the specified path.
//
// Parameters:
//   - name: The name of the entry to delete from the index.
//   - indexFilePath: The path to the index file.
func deleteFromIndex(name string, indexFilePath string) error {
	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		// If the index file does not exist, return false
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read index file: %w", err)
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to marshall index file: %w", err)
	}

	newEntries := make([]*pb.IndexListEntry, 0, len(index.Entries))
	for _, entry := range index.Entries {
		if entry.Name != name {
			newEntries = append(newEntries, entry)
		}
	}

	index.Entries = newEntries

	updatedData, err := proto.Marshal(&index)
	if err != nil {
		return fmt.Errorf("failed to marshal updated index: %w", err)
	}

	if err := os.WriteFile(indexFilePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated index to file: %w", err)
	}

	return nil
}

// deleteCopy deletes a copied file with the given name from the specified index path.
//
// Parameters:
//   - indexPath (string): The path to the index directory.
//   - name (string): The name of the copied file to delete.
func deleteCopy(indexPath, name string) error {
	copiedFilePath := path.Join(indexPath, addedCopiesSubDir, name)

	if _, err := os.Stat(copiedFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to check the stat of file %s: %v", copiedFilePath, err)
	}

	if err := os.Remove(copiedFilePath); err != nil {
		return fmt.Errorf("failed to remove copied file %s: %v", copiedFilePath, err)
	}

	return nil
}
