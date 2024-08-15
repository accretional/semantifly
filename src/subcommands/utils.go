package subcommands

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func readIndex(indexFilePath string, ignoreIfNotFound bool) (map[string]*pb.IndexListEntry, error) {
	index := &pb.Index{}
	data, err := os.ReadFile(indexFilePath)

	if err != nil {
		if os.IsNotExist(err) && ignoreIfNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	if err := proto.Unmarshal(data, index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index file: %w", err)
	}

	if !ignoreIfNotFound && (index == nil || len(index.Entries) == 0) {
		return nil, fmt.Errorf("empty index file")
	}

	indexMap := make(map[string]*pb.IndexListEntry)

	for _, entry := range index.Entries {
		indexMap[entry.Name] = entry
	}

	return indexMap, nil
}

func writeIndex(indexFilePath string, indexMap map[string]*pb.IndexListEntry) error {

	index := &pb.Index{
		Entries: make([]*pb.IndexListEntry, 0, len(indexMap)),
	}

	for _, entry := range indexMap {
		index.Entries = append(index.Entries, entry)
	}

	updatedData, err := proto.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal updated index: %w", err)
	}

	if err := os.WriteFile(indexFilePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated index to file: %w", err)
	}

	return nil
}
func makeCopy(indexPath string, ile *pb.IndexListEntry) error {

	dest := path.Join(indexPath, addedCopiesSubDir, ile.Name)

	srcFile, err := os.Open(ile.URI)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	ile.Content = string(content)
	ile.LastRefreshedTime = timestamppb.Now()

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

func parseDataType(str string) (pb.DataType, error) {
	val, ok := pb.DataType_value[strings.ToUpper(str)]
	if !ok {
		return pb.DataType_TEXT, fmt.Errorf("unknown data type: %s", str)
	}
	return pb.DataType(val), nil
}

func parseSourceType(str string) (pb.SourceType, error) {
	val, ok := pb.SourceType_value[strings.ToUpper(str)]
	if !ok {
		return pb.SourceType_LOCAL_FILE, fmt.Errorf("unknown source type: %s", str)
	}
	return pb.SourceType(val), nil
}
