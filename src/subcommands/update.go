package subcommands

import (
	"context"
	"fmt"
	"path"

	db "accretional.com/semantifly/database"
	fetch "accretional.com/semantifly/fetcher"
	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	search "accretional.com/semantifly/search"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UpdateArgs struct {
	Context    context.Context
	DBConn     db.PgxIface
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

	ile, err := updateIndex(indexMap, &u)
	if err != nil {
		fmt.Printf("Failed to update the index entry %s: %v", u.Name, err)
		return
	}

	err = search.CreateSearchDictionary(ile)
	if err != nil {
		fmt.Printf("Failed to update the search dictionary %s: %v", u, err)
		return
	}

	if u.UpdateCopy == "true" {

		content, err := fetch.FetchFromSource(ile.ContentMetadata.SourceType, u.DataURI)
		if err != nil {
			fmt.Printf("Failed to validate the URI %s: %v\n", u, err)
			return
		}

		ileWithContent := &pb.IndexListEntry{
			Name:            ile.Name,
			ContentMetadata: ile.ContentMetadata,
			Content:         string(content),
		}

		if err := makeCopy(u.IndexPath, ileWithContent); err != nil {
			fmt.Printf("Failed to update the copy of the source file: %v", err)
			return
		}
	}

	if err := writeIndex(indexFilePath, indexMap); err != nil {
		fmt.Printf("Failed to write to the index file: %v", err)
		return
	}

	if err := db.InsertRows(u.Context, u.DBConn, &pb.Index{Entries: []*pb.IndexListEntry{ile}}); err != nil {
		fmt.Printf("Failed to update the database: %v", err)
		return
	}

	fmt.Printf("Index %s updated successfully to URI %s\n", u.Name, u.DataURI)
}

func updateIndex(indexMap map[string]*pb.IndexListEntry, u *UpdateArgs) (*pb.IndexListEntry, error) {

	entry, exists := indexMap[u.Name]
	if !exists {
		return nil, fmt.Errorf("entry %s not found", u.Name)
	}

	entry.ContentMetadata.URI = u.DataURI

	if u.DataType != "" {
		if dataType, err := parseDataType(u.DataType); err != nil {
			return nil, fmt.Errorf("error in parsing DataType: %v", err)
		} else {
			entry.ContentMetadata.DataType = dataType
		}
	}

	if u.SourceType != "" {
		if sourceType, err := parseSourceType(u.SourceType); err != nil {
			return nil, fmt.Errorf("error in parsing SourceType: %v", err)
		} else {
			entry.ContentMetadata.SourceType = sourceType
		}
	}

	entry.LastRefreshedTime = timestamppb.Now()
	indexMap[u.Name] = entry

	return entry, nil
}
