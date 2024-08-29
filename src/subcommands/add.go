package subcommands

import (
	"fmt"
	"io"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	search "accretional.com/semantifly/search"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubcommandAdd(a *pb.AddRequest, indexPath string, w io.Writer) error {
	indexFilePath := path.Join(indexPath, indexFile)
	indexMap, err := readIndex(indexFilePath, true)
	if err != nil {
		return fmt.Errorf("Failed to read the index file: %v", err)
	}

	for _, u := range a.DataUris {
		if indexMap[u] != nil {
			fmt.Fprintf(w, "File %s has already been added. Skipping without refresh.\n", u)
			continue
		}

		ile := &pb.IndexListEntry{
			Name: u,

			ContentMetadata: &pb.ContentMetadata{
				URI:        u,
				DataType:   a.DataType,
				SourceType: a.SourceType,
			},

			FirstAddedTime: timestamppb.Now(),
		}

		if a.MakeCopy {
			err = makeCopy(indexPath, ile)
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
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("Failed to write to the index file: %v", err)
	}

	return nil
}
