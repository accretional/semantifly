package subcommands

import (
	"fmt"
	"io"
	"path"

	fetch "accretional.com/semantifly/fetcher"
)

type GetArgs struct {
	IndexPath string
	Name      string
}

func Get(g GetArgs, w io.Writer) (string, error) {
	indexFilePath := path.Join(g.IndexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		return "", fmt.Errorf("failed to read the index file: %v", err)
	}

	targetEntry := indexMap[g.Name]

	if targetEntry == nil {
		fmt.Fprintf(w, "entry '%s' not found in index file %s\n", g.Name, indexFilePath)
		return "", fmt.Errorf("entry '%s' not found in index file %s", g.Name, indexFilePath)
	}

	if targetEntry.Content != "" {
		fmt.Fprintln(w, targetEntry.Content)
	} else {
		content, err := fetchFromCopy(g.IndexPath, g.Name)
		if content == nil {
			if err != nil {
				fmt.Fprintf(w, "failed to read content from copy: %v. Fetching from the source.\n", err)
			}

			content, err = fetch.FetchFromSource(targetEntry.SourceType, targetEntry.URI)
			if err != nil {
				return "", fmt.Errorf("failed to read content from source: %v", err)
			}
		}

		fmt.Fprintln(w, string(content))
		return string(content), nil
	}

	return targetEntry.Content, nil
}
