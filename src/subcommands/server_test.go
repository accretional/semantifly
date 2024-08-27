package subcommands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const testDir = "./test_semantifly"
const testIndexPath = "test"

func setupTestEnvironment(t *testing.T) func() {
	// Create test directory
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create some test files
	testFiles := []string{"test_file1.txt", "test_file2.txt"}
	for _, file := range testFiles {
		filePath := filepath.Join(testDir, file)
		err := os.WriteFile(filePath, []byte("This is a test file for Semantifly"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	return func() {
		// Clean up test directory
		os.RemoveAll(testDir)
	}
}

func setupServerAndClient(t *testing.T) (pb.SemantiflyClient, func()) {
	// Set up the server
	stopServer := startTestServer()

	// Set up the client
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient("localhost:50051", opts...)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	client := pb.NewSemantiflyClient(conn)

	cleanup := func() {
		conn.Close()
		stopServer()
	}

	return client, cleanup
}

func startTestServer() func() {
	go executeStartServer([]string{"--semantifly_dir", testDir})

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	return func() {
		// Implement server shutdown logic here if needed
	}
}

func TestServerAdd(t *testing.T) {
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.AddRequest{
		IndexPath:  testIndexPath,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt")},
	}

	resp, err := client.Add(ctx, req)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("Add was not successful: %s", resp.Message)
	}
}

func TestServerGet(t *testing.T) {
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	// Add a file first
	addCtx, addCancel := context.WithTimeout(context.Background(), time.Second)
	_, err := client.Add(addCtx, &pb.AddRequest{
		IndexPath:  testIndexPath,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt")},
	})
	addCancel()
	if err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.GetRequest{
		IndexPath: testIndexPath,
		Name:      "test_file1.txt",
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("Get was not successful: %s", resp.Message)
	}
	if resp.Content == "" {
		t.Errorf("Get returned empty content")
	}
}

func TestServerDelete(t *testing.T) {
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	// Add a file first
	addCtx, addCancel := context.WithTimeout(context.Background(), time.Second)
	_, err := client.Add(addCtx, &pb.AddRequest{
		IndexPath:  testIndexPath,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt")},
	})
	addCancel()
	if err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.DeleteRequest{
		IndexPath:  testIndexPath,
		DeleteCopy: true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt")},
	}

	resp, err := client.Delete(ctx, req)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("Delete was not successful: %s", resp.Message)
	}
}

func TestServerUpdate(t *testing.T) {
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	// Add a file first
	addCtx, addCancel := context.WithTimeout(context.Background(), time.Second)
	_, err := client.Add(addCtx, &pb.AddRequest{
		IndexPath:  testIndexPath,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt")},
	})
	addCancel()
	if err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.UpdateRequest{
		IndexPath:  testIndexPath,
		Name:       "test_file1.txt",
		DataType:   "text",
		SourceType: "local_file",
		UpdateCopy: true,
		DataUri:    filepath.Join(testDir, "test_file2.txt"),
	}

	resp, err := client.Update(ctx, req)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("Update was not successful: %s", resp.Message)
	}
}

func TestServerLexicalSearch(t *testing.T) {
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	// Add files first
	addCtx, addCancel := context.WithTimeout(context.Background(), time.Second)
	_, err := client.Add(addCtx, &pb.AddRequest{
		IndexPath:  testIndexPath,
		DataType:   "text",
		SourceType: "local_file",
		MakeCopy:   true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt"), filepath.Join(testDir, "test_file2.txt")},
	})
	addCancel()
	if err != nil {
		t.Fatalf("Failed to add test files: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.LexicalSearchRequest{
		IndexPath:  testIndexPath,
		SearchTerm: "test",
		TopN:       5,
	}

	resp, err := client.LexicalSearch(ctx, req)
	if err != nil {
		t.Fatalf("LexicalSearch failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("LexicalSearch was not successful: %s", resp.Message)
	}
	if len(resp.Results) == 0 {
		t.Errorf("LexicalSearch returned no results")
	}
}
