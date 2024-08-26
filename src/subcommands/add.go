package subcommands

import (
	"fmt"
	"io"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	search "accretional.com/semantifly/search"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AddArgs struct {
	IndexPath  string
	DataType   string
	SourceType string
	MakeCopy   bool
	DataURIs   []string
}

func Add(a AddArgs, w io.Writer) error {
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

	for _, u := range a.DataURIs {
		if indexMap[u] != nil {
			fmt.Fprintf(w, "File %s has already been added. Skipping without refresh.\n", u)
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
				fmt.Fprintf(w, "Failed to make a copy for %s: %v. Skipping.\n", u, err)
				continue
			}
		}

		err = search.CreateSearchDictionary(ile)
		if err != nil {
			fmt.Fprintf(w, "File %s failed to create search dictionary with err: %s. Skipping.\n", u, err)
			continue
		}

		indexMap[ile.Name] = ile
		fmt.Fprintf(w, "Index %s added successfully.\n", u)
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("Failed to write to the index file: %v", err)
	}

	return nil
}
