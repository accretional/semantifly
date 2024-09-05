package subcommands

import (
	"fmt"
	"io"
	"os"
	"path"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	search "accretional.com/semantifly/search"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubcommandAdd(a *pb.AddRequest, indexPath string, w io.Writer) error {
	if err := createDirectoriesIfNotExist(indexPath); err != nil {
		return fmt.Errorf("Failed to create directories: %v", err)
	}

	indexFilePath := path.Join(indexPath, indexFile)
	indexMap, err := readIndex(indexFilePath, true)
	if err != nil {
		return fmt.Errorf("Failed to read the index file: %v", err)
	}

	if indexMap[a.AddedMetadata.URI] != nil {
		return fmt.Errorf("File %s has already been added. Skipping without refresh.\n", a.AddedMetadata.URI)
	}

	ile := &pb.IndexListEntry{
		Name:            a.AddedMetadata.URI,
		ContentMetadata: a.AddedMetadata,
		FirstAddedTime:  timestamppb.Now(),
	}

	if a.MakeCopy {
		err = makeCopy(indexPath, ile)
		if err != nil {
			fmt.Fprintf(w, "Failed to make a copy for %s: %v. Skipping.\n", a.AddedMetadata.URI, err)
		}
	}

	err = search.CreateSearchDictionary(ile)
	if err != nil {
		fmt.Fprintf(w, "File %s failed to create search dictionary with err: %s. Skipping.\n", a, err)
	}

	indexMap[ile.Name] = ile

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("Failed to write to the index file: %v", err)
	}

	return nil
}

func createDirectoriesIfNotExist(dir string) error {
	if _, err := os.ReadDir(dir); err != nil {
		fmt.Printf("No existing directory detected. Creating new directory at %s\n", dir)
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("Failed to create directory at %s: %s", dir, err)
		}
	}
	return nil
}
