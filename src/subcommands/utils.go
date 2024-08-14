package subcommands

import (
	"fmt"
	"os"
	"strings"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/proto"
)

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
