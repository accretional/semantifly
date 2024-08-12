package subcommands

import (
	"fmt"
	"os"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

type GetArgs struct {
	IndexPath string
	Name      string
}

func Get(g GetArgs) {
	indexFilePath := path.Join(g.IndexPath, indexFile)

	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		fmt.Printf("failed to read index file: %v\n", err)
		return
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		fmt.Printf("failed to unmarshal index file: %v\n", err)
		return
	}

	var targetEntry *pb.IndexListEntry
	for _, entry := range index.Entries {
		if entry.Name == g.Name {
			targetEntry = entry
			break
		}
	}

	if targetEntry == nil {
		fmt.Printf("file '%s' not found in the index\n", g.Name)
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
