package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

func FetchFromSource(sourceType pb.SourceType, uri string) ([]byte, error) {
	var fetch func(string) ([]byte, error)

	switch sourceType {
	case pb.SourceType_LOCAL_FILE:
		fetch = fetchFromFile

	case pb.SourceType_WEBPAGE:
		fetch = fetchFromWebpage

	default:
		return nil, fmt.Errorf("invalid sourceType argument")
	}

	content, err := fetch(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from source %s: %w", uri, err)
	}

	return content, nil
}

func fetchFromFile(uri string) ([]byte, error) {
	fmt.Println("uri")
	fmt.Println(uri)
	f, err := os.Stat(uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file does not exist")
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if f.IsDir() {
		return nil, fmt.Errorf("cannot add directory %s as file. Try adding as a directory instead", uri)
	}

	if !f.Mode().IsRegular() {
		return nil, fmt.Errorf("file %s is not a regular file and cannot be added", uri)
	}

	srcFile, err := os.Open(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	return content, nil
}

func fetchFromWebpage(uri string) ([]byte, error) {

	// Using a context with timeout for HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch web page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("web page returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read web page content: %v", err)
	}

	return body, nil
}
