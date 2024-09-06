package subcommands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	db "accretional.com/semantifly/database"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)


func SubcommandDelete(ctx context.Context, conn *db.PgxIface, d *pb.DeleteRequest, indexPath string, w io.Writer) error {
	indexFilePath := path.Join(indexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		return fmt.Errorf("failed to read the index file: %v", err)
	}

	for _, uri := range d.Names {

		if _, present := indexMap[uri]; !present {
			fmt.Fprintf(w, "Entry %s not found in index file %s, skipping\n", uri, indexFilePath)
			continue
		}

		delete(indexMap, uri)

		if d.DeleteCopy {
			if err := deleteCopy(indexPath, uri, w); err != nil {
				fmt.Fprintf(w, "Failed to delete copy of file %s with err: %s, skipping", uri, err)
			}
		}
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("failed to write to the index file: %v", err)
	}

	if err := db.DeleteRows(ctx, conn, d.Names); err != nil {
		return fmt.Errorf("failed to delete from the database: %v", err)
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
