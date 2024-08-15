package subcommands

import (
	"fmt"
	"path"
)

type GetArgs struct {
	IndexPath string
	Name      string
}

func Get(g GetArgs) {
	indexFilePath := path.Join(g.IndexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		fmt.Printf("Failed to read the index file: %v", err)
		return
	}

	targetEntry := indexMap[g.Name]

	if targetEntry == nil {
		fmt.Printf("file '%s' not found in index file %s\n", g.Name, indexFilePath)
		return
	}

	if targetEntry.Content != "" {
		fmt.Println(targetEntry.Content)
	} else {
		content, err := fetchFromCopy(g.IndexPath, g.Name)
		if content == nil {
			if err != nil {
				fmt.Printf("Failed to read content from copy: %v. Fetching from the source.\n", err)
			}

			content, err = fetchFromSource(targetEntry.SourceType, targetEntry.URI)
			if err != nil {
				fmt.Printf("Failed to read content from source: %v.\n", err)
				return
			}
		}

		fmt.Println(string(content))
	}
}
