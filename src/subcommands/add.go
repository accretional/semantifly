package subcommands

import (
	"fmt"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"

	search "accretional.com/semantifly/search"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubcommandAdd(a *pb.AddRequest) error {
	dataType, err := parseDataType(a.DataType)
	if err != nil {
		return fmt.Errorf("Error in parsing DataType: %v\n", err)
	}

	sourceType, err := parseSourceType(a.SourceType)
	if err != nil {
		return fmt.Errorf("Error in parsing SourceType: %v\n", err)
	}

	indexFilePath := path.Join(a.IndexPath, indexFile)
	indexMap, err := readIndex(indexFilePath, true)
	if err != nil {
		return fmt.Errorf("Failed to read the index file: %v", err)
	}

	for _, u := range a.DataUris {
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

		if a.MakeCopy {
			err = makeCopy(a.IndexPath, ile)
			if err != nil {
				fmt.Printf("Failed to make a copy for %s: %v. Skipping.\n", u, err)
				continue
			}
		}

		err = search.CreateSearchDictionary(ile)
		if err != nil {
			fmt.Printf("File %s failed to create search dictionary with err: %s. Skipping.\n", u, err)
			continue
		}

		indexMap[ile.Name] = ile
		fmt.Printf("Index %s added successfully.\n", u)
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("Failed to write to the index file: %v", err)
	}

	return nil
}
