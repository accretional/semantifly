package database

import (
	"context"
	"fmt"
	"os"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"github.com/jackc/pgx/v5"
)

func AddOrUpdateDatabase(indexMap map[string]*pb.IndexListEntry) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to initialise the database schema: %v", err)
	}

	index := &pb.Index{
		Entries: make([]*pb.IndexListEntry, 0, len(indexMap)),
	}

	for _, entry := range indexMap {
		index.Entries = append(index.Entries, entry)
	}

	err = insertRows(ctx, conn, index)
	if err != nil {
		return fmt.Errorf("failed to insert rows: %v", err)
	}

	if err := conn.Close(ctx); err != nil {
		return fmt.Errorf("failed to close the connection after insertion: %v", err)
	}

	return nil
}

func DeleteFromDatabase(name string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to initialise the database schema: %v", err)
	}

	err = deleteRow(ctx, conn, name)
	if err != nil {
		return fmt.Errorf("failed to insert rows: %v", err)
	}

	if err := conn.Close(ctx); err != nil {
		return fmt.Errorf("failed to close the connection after insertion: %v", err)
	}

	return nil
}

func GetFromDatabase(name string) (*pb.IndexListEntry, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	err = initializeDatabaseSchema(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise the database schema: %v", err)
	}

	metadata, err := getContentMetadata(ctx, conn, name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert rows: %v", err)
	}

	if err := conn.Close(ctx); err != nil {
		return nil, fmt.Errorf("failed to close the connection after insertion: %v", err)
	}

	ile := &pb.IndexListEntry{
		Name:            name,
		ContentMetadata: metadata,
	}

	return ile, nil
}
