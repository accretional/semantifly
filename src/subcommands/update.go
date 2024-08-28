package subcommands

import (
	"fmt"
	"path"

	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UpdateArgs struct {
	Name       string
	IndexPath  string
	DataType   string
	SourceType string
	UpdateCopy string
	DataURI    string
}

func Update(u UpdateArgs) {
	indexFilePath := path.Join(u.IndexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		fmt.Printf("Failed to read the index file: %v", err)
		return
	}

	if err := updateIndex(indexMap, &u); err != nil {
		fmt.Printf("Failed to update the index entry %s: %v", u.Name, err)
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		fmt.Printf("Failed to write to the index file: %v", err)
		return
	}

	if u.UpdateCopy == "true" {

		sourceType, err := parseSourceType(u.SourceType)
		if err != nil {
			fmt.Printf("Invalid source type: %v", err)
			return
		}

		content, err := fetch.FetchFromSource(sourceType, u.DataURI)

		if err != nil {
			fmt.Printf("Failed to validate the URI %s: %v\n", u, err)
			return
		}

		ile := &pb.IndexListEntry{
			Name: u.Name,
			ContentMetadata: &pb.ContentMetadata{
				URI:        u.DataURI,
				DataType:   pb.DataType(pb.DataType_value[u.DataType]),
				SourceType: pb.SourceType(pb.SourceType_value[u.SourceType]),
			},
			Content: string(content),
		}

		if err := makeCopy(u.IndexPath, ile); err != nil {
			fmt.Printf("Failed to update the copy of the source file: %v", err)
			return
		}
	}

	fmt.Printf("Index %s updated successfully to URI %s\n", u.Name, u.DataURI)
}

func updateIndex(indexMap map[string]*pb.IndexListEntry, u *UpdateArgs) error {

	entry, exists := indexMap[u.Name]
	if !exists {
		return fmt.Errorf("entry %s not found", u.Name)
	}

	entry.ContentMetadata.URI = u.DataURI

	if u.DataType != "" {
		if dataType, err := parseDataType(u.DataType); err != nil {
			return fmt.Errorf("error in parsing DataType: %v", err)
		} else {
			entry.ContentMetadata.DataType = dataType
		}
	}

	if u.SourceType != "" {
		if sourceType, err := parseSourceType(u.SourceType); err != nil {
			return fmt.Errorf("error in parsing SourceType: %v", err)
		} else {
			entry.ContentMetadata.SourceType = sourceType
		}
	}

	entry.LastRefreshedTime = timestamppb.Now()
	indexMap[u.Name] = entry

	u.SourceType = pb.SourceType_name[int32(entry.ContentMetadata.SourceType)]
	u.DataType = pb.DataType_name[int32(entry.ContentMetadata.DataType)]

	return nil
}
