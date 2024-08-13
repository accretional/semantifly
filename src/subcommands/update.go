package subcommands

import (
	"fmt"
	"os"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UpdateArgs struct {
	Name       string
	IndexPath  string
	DataType   pb.DataType
	SourceType pb.SourceType
	UpdateCopy string
	DataURI    string
}

func Update(u UpdateArgs) {
	indexFilePath := path.Join(u.IndexPath, indexFile)
	var index pb.Index

	ile := &pb.IndexListEntry{
		Name:              u.Name,
		URI:               u.DataURI,
		DataType:          u.DataType,
		SourceType:        u.SourceType,
		LastRefreshedTime: timestamppb.Now(),
	}

	if err := readIndex(indexFilePath, &index); err != nil {
		fmt.Printf("Failed to read the index file: %v", err)
	}

	if entryFound:= updateIndex(&index, ile); !entryFound {
		fmt.Printf("Entry %s not found in the index file %s", u.Name, indexFilePath)
	}

	if err := writeIndex(indexFilePath, &index); err != nil {
		fmt.Printf("Failed to write to the index file: %v", err)
	}

	if u.UpdateCopy == "true" {
		if err := updateCopy(u.Name, path.Join(u.IndexPath, addedCopiesSubDir, u.Name)); err != nil {
			fmt.Printf("Failed to update the copy of the source file: %v", err)
		}
	}
}

func readIndex(indexFilePath string, index *pb.Index, ignoreIfNotFound ...bool) error {
	data, err := os.ReadFile(indexFilePath)
	if err != nil {
		if os.IsNotExist(err) && (len(ignoreIfNotFound) > 0 && ignoreIfNotFound[0]) {
			return nil
		}
		return fmt.Errorf("failed to read index file: %w", err)
	}

	if err := proto.Unmarshal(data, index); err != nil {
		return fmt.Errorf("failed to marshall index file: %w", err)
	}

	return nil
}

func updateIndex(index *pb.Index, ile *pb.IndexListEntry) bool{
	for i, entry := range index.Entries {
		if entry.Name == ile.Name {
			if ile.URI != "" {
				index.Entries[i].URI = ile.URI
			}
			if ile.DataType != pb.DataType_DATA_TYPE_UNSPECIFIED {
				index.Entries[i].DataType = ile.DataType
			}
			if ile.SourceType != pb.SourceType_SOURCE_TYPE_UNSPECIFIED {
				index.Entries[i].SourceType = ile.SourceType
			}
			index.Entries[i].LastRefreshedTime = ile.LastRefreshedTime

			return true
		}
	}

	return false
}

func writeIndex(indexFilePath string, index *pb.Index) error {
	updatedData, err := proto.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal updated index: %w", err)
	}
	if err := os.WriteFile(indexFilePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated index to file: %w", err)
	}

	return nil
}
