package subcommands

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
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
	var index pb.Index

	if err := readIndex(indexFilePath, &index); err != nil {
		fmt.Printf("Failed to read the index file: %v", err)
		return
	}

	if err := updateIndex(&index, &u); err != nil {
		fmt.Printf("Failed to update the index entry %s: %v", u.Name, err)
	}

	if err := writeIndex(indexFilePath, &index); err != nil {
		fmt.Printf("Failed to write to the index file: %v", err)
		return
	}

	if u.UpdateCopy == "true" {
		if err := updateCopy(&u); err != nil {
			fmt.Printf("Failed to update the copy of the source file: %v", err)
			return
		}
	}

	fmt.Printf("Index %s updated successfully to URI %s\n", u.Name, u.DataURI)
}

func updateIndex(index *pb.Index, u *UpdateArgs) error {
	for i, entry := range index.Entries {
		if entry.Name == u.Name {
			index.Entries[i].URI = u.DataURI

			if u.DataType != "" {
				if dataType, err := parseDataType(u.DataType); err != nil {
					return fmt.Errorf("error in parsing DataType: %v", err)
				} else {
					index.Entries[i].DataType = dataType
				}
			}

			if u.SourceType != "" {
				if sourceType, err := parseSourceType(u.SourceType); err != nil {
					return fmt.Errorf("error in parsing SourceType: %v", err)
				} else {
					index.Entries[i].SourceType = sourceType
				}
			}

			index.Entries[i].LastRefreshedTime = timestamppb.Now()

			u.SourceType = pb.SourceType_name[int32(index.Entries[i].SourceType)]
			u.DataType = pb.DataType_name[int32(index.Entries[i].DataType)]

			return nil
		}
	}

	return fmt.Errorf("entry %s not found", u.Name)
}

func updateCopy(u *UpdateArgs) error {

	dest := path.Join(u.IndexPath, addedCopiesSubDir, u.Name)

	srcFile, err := os.Open(u.DataURI)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	ile := &pb.IndexListEntry{
		Name:              u.Name,
		URI:               u.DataURI,
		DataType:          pb.DataType(pb.DataType_value[u.DataType]),
		SourceType:        pb.SourceType(pb.SourceType_value[u.SourceType]),
		Content:           string(content),
		LastRefreshedTime: timestamppb.Now(),
	}

	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("failed to create destination dir %s: %w", dir, err)
	}

	data, err := proto.Marshal(ile)
	if err != nil {
		return fmt.Errorf("failed to marshal IndexListEntry: %w", err)
	}

	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}
