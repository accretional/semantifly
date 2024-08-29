package subcommands

import (
	"fmt"
	"io"
	"path"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func SubcommandGet(g *pb.GetRequest, indexPath string, w io.Writer) (string, error) {
	indexFilePath := path.Join(indexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		return "", fmt.Errorf("failed to read the index file: %v", err)
	}

	targetEntry := indexMap[g.Name]

	if targetEntry == nil {
		fmt.Fprintf(w, "entry '%s' not found in index file %s\n", g.Name, indexFilePath)
		return "", fmt.Errorf("entry '%s' not found in index file %s", g.Name, indexFilePath)
	}

	if targetEntry.Content == "" {
		content, err := fetchFromCopy(indexPath, g.Name)
		if content == nil {
			if err != nil {
				fmt.Fprintf(w, "failed to read content from copy: %v. Fetching from the source.\n", err)
			}

			content, err = fetch.FetchFromSource(targetEntry.ContentMetadata.SourceType, targetEntry.ContentMetadata.URI)
			if err != nil {
				return "", fmt.Errorf("failed to read content from source: %v", err)
			}
		}

		return string(content), nil
	}

	return targetEntry.Content, nil
}
