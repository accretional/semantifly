package subcommands

import (
	"errors"
	"fmt"
	"os"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const addedCopiesSubDir = "add_cache"
const indexFile = "index.list"

type AddArgs struct {
	IndexPath  string
	DataType   string
	SourceType string
	MakeCopy   bool
	DataURIs   []string
}

func Add(a AddArgs) {

	dataType, err := parseDataType(a.DataType)
	if err != nil {
		fmt.Printf("error in parsing DataType: %v", err)
		return
	}

	sourceType, err := parseSourceType(a.SourceType)
	if err != nil {
		fmt.Printf("Invalid source type: %v", err)
		return
	}

	switch sourceType {
	case pb.SourceType_LOCAL_FILE:

		indexFilePath := path.Join(a.IndexPath, indexFile)
		indexMap, err := readIndex(indexFilePath, true)
		if err != nil {
			fmt.Printf("Failed to read the index file: %v", err)
			return
		}

		for i, u := range a.DataURIs {
			f, err := os.Stat(u)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Failed to add file %s at input list index %v: file does not exist\n", u, i)
				continue
			}
			if f.IsDir() {
				fmt.Printf("Cannot add directory %s as file. Try adding as a directory instead. Skipping.\n", u)
				continue
			}
			if !f.Mode().IsRegular() {
				fmt.Printf("File %s is not a regular file and cannot be added. Skipping.\n", u)
				continue
			}

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
					fmt.Printf("File %s failed to copy with err: %s. Skipping.\n", u, err)
					continue
				}
			}

			fmt.Printf("Index %s added successfully.\n", u)
		}

		if err := writeIndex(indexFilePath, indexMap); err != nil {
			fmt.Printf("Failed to write to the index file: %v", err)
			return
		}

	default:
		fmt.Printf("Invalid 'add' SourceType subsubcommand: %s\n", a.SourceType)
		os.Exit(1)
	}
}
