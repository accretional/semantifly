package subcommands

import (
	"fmt"
	"io"
	"path"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubcommandUpdate(u *pb.UpdateRequest, indexPath string, w io.Writer) error {
	indexFilePath := path.Join(indexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		return fmt.Errorf("failed to read the index file: %v", err)
	}

	if err := updateIndex(indexMap, u); err != nil {
		return fmt.Errorf("failed to update the index entry %s: %v", u.Name, err)
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("failed to write to the index file: %v", err)
	}

	if u.UpdateCopy {
		content, err := fetch.FetchFromSource(u.FileData.SourceType, u.FileData.URI)

		if err != nil {
			return fmt.Errorf("failed to validate the URI %s: %v\n", u, err)
		}

		ile := &pb.IndexListEntry{
			Name:            u.Name,
			ContentMetadata: u.FileData,
			Content:         string(content),
		}

		if err := makeCopy(indexPath, ile); err != nil {
			return fmt.Errorf("failed to update the copy of the source file: %v", err)
		}
	}
	return nil
}

func updateIndex(indexMap map[string]*pb.IndexListEntry, u *pb.UpdateRequest) error {

	entry, exists := indexMap[u.Name]
	if !exists {
		return fmt.Errorf("entry %s not found", u.Name)
	}

	entry.ContentMetadata = u.FileData

	entry.LastRefreshedTime = timestamppb.Now()
	indexMap[u.Name] = entry

	return nil
}
