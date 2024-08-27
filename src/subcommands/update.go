package subcommands

import (
	"fmt"
	"io"
	"path"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubcommandUpdate(u *pb.UpdateRequest, w io.Writer) error {
	indexFilePath := path.Join(u.IndexPath, indexFile)

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

		sourceType, err := parseSourceType(u.SourceType)
		if err != nil {
			return fmt.Errorf("invalid source type: %v", err)
		}

		content, err := fetch.FetchFromSource(sourceType, u.DataUri)

		if err != nil {
			return fmt.Errorf("failed to validate the URI %s: %v\n", u, err)
		}

		ile := &pb.IndexListEntry{
			Name:       u.Name,
			URI:        u.DataUri,
			DataType:   pb.DataType(pb.DataType_value[u.DataType]),
			SourceType: pb.SourceType(pb.SourceType_value[u.SourceType]),
			Content:    string(content),
		}

		if err := makeCopy(u.IndexPath, ile); err != nil {
			return fmt.Errorf("failed to update the copy of the source file: %v", err)
		}
	}

	fmt.Fprintf(w, "Index %s updated successfully to URI %s\n", u.Name, u.DataUri)
	return nil
}

func updateIndex(indexMap map[string]*pb.IndexListEntry, u *pb.UpdateRequest) error {

	entry, exists := indexMap[u.Name]
	if !exists {
		return fmt.Errorf("entry %s not found", u.Name)
	}

	entry.URI = u.DataUri

	if u.DataType != "" {
		if dataType, err := parseDataType(u.DataType); err != nil {
			return fmt.Errorf("error in parsing DataType: %v", err)
		} else {
			entry.DataType = dataType
		}
	}

	if u.SourceType != "" {
		if sourceType, err := parseSourceType(u.SourceType); err != nil {
			return fmt.Errorf("error in parsing SourceType: %v", err)
		} else {
			entry.SourceType = sourceType
		}
	}

	entry.LastRefreshedTime = timestamppb.Now()
	indexMap[u.Name] = entry

	u.SourceType = pb.SourceType_name[int32(entry.SourceType)]
	u.DataType = pb.DataType_name[int32(entry.DataType)]

	return nil
}
