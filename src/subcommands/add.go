package subcommands

import (
	"fmt"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AddArgs struct {
	IndexPath  string
	DataType   string
	SourceType string
	MakeCopy   bool
	DataURIs   []string
}

func Add(a AddArgs) {

	dataType, err := parseDataType(a.DataType)
	if err != nil {
		fmt.Printf("error in parsing DataType: %v", err)
		return
	}

	sourceType, err := parseSourceType(a.SourceType)
	if err != nil {
		fmt.Printf("Invalid source type: %v", err)
		return
	}

	indexFilePath := path.Join(a.IndexPath, indexFile)
	indexMap, err := readIndex(indexFilePath, true)
	if err != nil {
		fmt.Printf("Failed to read the index file: %v", err)
		return
	}

	for _, u := range a.DataURIs {

		if indexMap[u] != nil {
			fmt.Printf("File %s has already been added. Skipping without refresh.\n", u)
			continue
		}

		ile := &pb.IndexListEntry{
			Name:           u,
			URI:            u,
			DataType:       dataType,
			SourceType:     sourceType,
			FirstAddedTime: timestamppb.Now(),
		}

		indexMap[ile.Name] = ile

		if a.MakeCopy {
			err = makeCopy(a.IndexPath, ile)
			if err != nil {
				fmt.Printf("Failed to make a copy for %s: %v. Skipping.\n", u, err)
				continue
			}
		}

		fmt.Printf("Index %s added successfully.\n", u)
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		fmt.Printf("Failed to write to the index file: %v", err)
		return
	}
}
