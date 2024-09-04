package subcommands

import (
	"context"
	"fmt"
	"os"

	db "accretional.com/semantifly/database"
	fetch "accretional.com/semantifly/fetcher"
)

type GetArgs struct {
	Context   context.Context
	DBConn    db.PgxIface
	IndexPath string
	Name      string
}

func Get(g GetArgs) {

	content, err := fetchFromCopy(g.IndexPath, g.Name)
	if content == nil {
		if os.IsNotExist(err) {
			fmt.Printf("Copy file does not exist. Fetching from source.\n")
		} else if err != nil {
			fmt.Printf("Failed to read content from copy: %v. Fetching from the source.\n", err)
		}

		targetMetadata, err := db.GetContentMetadata(g.Context, g.DBConn, g.Name)
		if err != nil {
			fmt.Printf("failed to get entry from database: %v\n", err)
			return
		} else if targetMetadata == nil {
			fmt.Printf("entry '%s' not found in database\n", g.Name)
			return
		}

		content, err = fetch.FetchFromSource(targetMetadata.SourceType, targetMetadata.URI)
		if err != nil {
			fmt.Printf("Failed to read content from source: %v.\n", err)
			return
		}
	}

	fmt.Println(string(content))
}
