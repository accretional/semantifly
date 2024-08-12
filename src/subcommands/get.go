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
	FileName  string
}

func Get(g GetArgs) error {
	indexFilePath := path.Join(g.IndexPath, indexFile)

	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	var index pb.Index
	if err := proto.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to unmarshal index file: %w", err)
	}

	var targetEntry *pb.IndexListEntry
	for _, entry := range index.Entries {
		if entry.Name == g.FileName {
			targetEntry = entry
			break
		}
	}

	if targetEntry == nil {
		return fmt.Errorf("file '%s' not found in the index", g.FileName)
	}

	if targetEntry.Content != "" {
		fmt.Println(targetEntry.Content)
	} else {
		content, err := os.ReadFile(targetEntry.URI)
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %w", g.FileName, err)
		}
		fmt.Println(string(content))
	}

	return nil
}