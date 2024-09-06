package subcommands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	db "accretional.com/semantifly/database"
	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

// TODO:
//
//	Get rid of index file. The database call does the same functionality
func SubcommandGet(ctx context.Context, conn *db.PgxIface, g *pb.GetRequest, indexPath string, w io.Writer) (string, *pb.ContentMetadata, error) {

	var targetMetadata pb.ContentMetadata

	switch g.IndexSource {
	case pb.IndexSource_INDEX_FILE:
		indexFilePath := path.Join(indexPath, indexFile)

		indexMap, err := readIndex(indexFilePath, false)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read the index file: %v", err)
		}

		targetEntry := indexMap[g.Name]

		if targetEntry == nil {
			fmt.Fprintf(w, "entry '%s' not found in index file %s\n", g.Name, indexFilePath)
			return "", nil, fmt.Errorf("entry '%s' not found in index file %s", g.Name, indexFilePath)
		}

		if targetEntry.Content != "" {
			return targetEntry.Content, targetEntry.ContentMetadata, nil
		}

		targetMetadata = *targetEntry.GetContentMetadata()

	case pb.IndexSource_DATABASE:

		dbMetadata, err := db.GetContentMetadata(ctx, conn, g.Name)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get entry from database: %v", err)
		} else if dbMetadata == nil {
			return "", nil, fmt.Errorf("entry '%s' not found in database", g.Name)

		}

		targetMetadata = *dbMetadata
	}

	content, err := fetchFromCopy(indexPath, g.Name)
	if content != nil {
		return string(content), &targetMetadata, nil
	} else if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(w, "failed to read content from copy: %v. Fetching from the source.\n", err)
	}

	content, err = fetch.FetchFromSource(targetMetadata.SourceType, targetMetadata.URI)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read content from source: %v", err)
	}

	return string(content), &targetMetadata, nil
}
