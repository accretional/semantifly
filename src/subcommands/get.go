package subcommands

import (
	"context"
	"fmt"
	"io"
	"path"

	db "accretional.com/semantifly/database"
	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

// TODO:  	Remove the dependency of index file in the get subcommand
//			The database call does the same functionality
func SubcommandGet(ctx context.Context, conn db.PgxIface, g *pb.GetRequest, indexPath string, w io.Writer) (string, *pb.ContentMetadata, error) {
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

	if targetEntry.Content == "" {
		content, err := fetchFromCopy(indexPath, g.Name)
		if content == nil {
			if err != nil {
				fmt.Fprintf(w, "failed to read content from copy: %v. Fetching from the source.\n", err)
			}

			targetMetadata, err := db.GetContentMetadata(ctx, conn, g.Name)
			if err != nil {
				return "", nil, fmt.Errorf("failed to get entry from database: %v\n", err)
			} else if targetMetadata == nil {
				return "", nil, fmt.Errorf("entry '%s' not found in database\n", g.Name)

			}

			content, err = fetch.FetchFromSource(targetMetadata.SourceType, targetMetadata.URI)
			if err != nil {
				return "", nil, fmt.Errorf("failed to read content from source: %v", err)
			}
		}

		return string(content), targetEntry.ContentMetadata, nil
	}

	return targetEntry.Content, targetEntry.ContentMetadata, nil
}
