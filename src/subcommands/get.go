package subcommands

import (
	"fmt"
	"os"
	"path"
)

type GetArgs struct {
	IndexPath string
	Name      string
}

func Get(g GetArgs) {
	indexFilePath := path.Join(g.IndexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, true)
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
		content, err := os.ReadFile(targetEntry.URI)
		if err != nil {
			fmt.Printf("failed to read file '%s': %v\n", g.Name, err)
			return
		}
		fmt.Println(string(content))
	}
}
