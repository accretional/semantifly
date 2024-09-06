package subcommands

import (
	"context"
	"fmt"
	"io"
	"path"

	db "accretional.com/semantifly/database"
	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	search "accretional.com/semantifly/search"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubcommandUpdate(ctx context.Context, conn *db.PgxIface, u *pb.UpdateRequest, indexPath string, w io.Writer) error {
	indexFilePath := path.Join(indexPath, indexFile)

	indexMap, err := readIndex(indexFilePath, false)
	if err != nil {
		return fmt.Errorf("failed to read the index file: %v", err)
	}

	if err := updateIndex(indexMap, u, w); err != nil {
		return fmt.Errorf("failed to update the index entry %s: %v", u.Name, err)
	}

	if u.UpdateCopy {
		content, err := fetch.FetchFromSource(u.UpdatedMetadata.SourceType, u.UpdatedMetadata.URI)

		if err != nil {
			return fmt.Errorf("failed to validate the URI %s: %v", u, err)
		}

		ile := &pb.IndexListEntry{
			Name:            u.Name,
			ContentMetadata: u.UpdatedMetadata,
			Content:         string(content),
		}

		if err := makeCopy(indexPath, ile); err != nil {
			return fmt.Errorf("failed to update the copy of the source file: %v", err)
		}
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		return fmt.Errorf("failed to write to the index file: %v", err)
	}

	if err := db.InsertRows(ctx, conn, &pb.Index{Entries: []*pb.IndexListEntry{indexMap[u.Name]}}); err != nil {
		return fmt.Errorf("failed to update the database: %v", err)
	}

	return nil
}

func updateIndex(indexMap map[string]*pb.IndexListEntry, u *pb.UpdateRequest, w io.Writer) error {
	entry, exists := indexMap[u.Name]
	if !exists {
		return fmt.Errorf("entry %s not found", u.Name)
	}

	entry.ContentMetadata = u.UpdatedMetadata

	entry.LastRefreshedTime = timestamppb.Now()

	err := search.CreateSearchDictionary(entry)
	if err != nil {
		fmt.Fprintf(w, "File %s failed to create search dictionary with err: %s. Skipping.\n", entry, err)
	}

	indexMap[u.Name] = entry

	return nil
}
