package subcommands

import (
	"fmt"
	"io"
	"os"
	"path"
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
func Delete(d DeleteArgs, w io.Writer) error {
	indexFilePath := path.Join(d.IndexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		return fmt.Errorf("failed to read the index file: %v", err)
	}

	for _, uri := range d.DataURIs {

		if _, present := indexMap[uri]; !present {
			fmt.Fprintf(w, "Entry %s not found in index file %s, skipping\n", uri, indexFilePath)
			continue
		}

		delete(indexMap, uri)

		fmt.Fprintf(w, "Index %s deleted successfully.\n", uri)

		if d.DeleteCopy {
			if err := deleteCopy(d.IndexPath, uri, w); err != nil {
				fmt.Fprintf(w, "Failed to delete copy of file %s with err: %s, skipping", uri, err)
			}
		}
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("failed to write to the index file: %v", err)
	}

	return nil
}

// deleteCopy deletes a copied file with the given name from the specified index path.
//
// Parameters:
//   - indexPath (string): The path to the index directory.
//   - name (string): The name of the copied file to delete.
func deleteCopy(indexPath, name string, w io.Writer) error {
	copiedFilePath := path.Join(indexPath, addedCopiesSubDir, name)

	if _, err := os.Stat(copiedFilePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(w, "No copy of %s found, skipping\n", name)
			return nil
		}

		return fmt.Errorf("failed to check the stat of file %s: %v", copiedFilePath, err)
	}

	if err := os.Remove(copiedFilePath); err != nil {
		return fmt.Errorf("failed to remove copied file %s: %v", copiedFilePath, err)
	}

	return nil
}
