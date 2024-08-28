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
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

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

	}
}

func TestServerCommands(t *testing.T) {
	// Start server
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	// Add
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

	// Get
	getCtx, getCancel := context.WithTimeout(context.Background(), time.Second)
	defer getCancel()

	getReq := &pb.GetRequest{
		IndexPath: testIndexPath,
		Name:      filepath.Join(testDir, "test_file1.txt"),
	}

	getResp, err := client.Get(getCtx, getReq)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !getResp.Success {
		t.Errorf("Get was not successful: %s", getResp.Message)
	}
	if getResp.Content == "" {
		t.Errorf("Get returned empty content")
	}

	// Update
	updCtx, updCancel := context.WithTimeout(context.Background(), time.Second)
	defer updCancel()

	updReq := &pb.UpdateRequest{
		IndexPath:  testIndexPath,
		Name:       filepath.Join(testDir, "test_file1.txt"),
		DataType:   "text",
		SourceType: "local_file",
		UpdateCopy: true,
		DataUri:    filepath.Join(testDir, "test_file2.txt"),
	}

	updResp, err := client.Update(updCtx, updReq)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !updResp.Success {
		t.Errorf("Update was not successful: %s", updResp.Message)
	}

	// Search
	searchCtx, searchCancel := context.WithTimeout(context.Background(), time.Second)
	defer searchCancel()

	searchReq := &pb.LexicalSearchRequest{
		IndexPath:  testIndexPath,
		SearchTerm: "test",
		TopN:       5,
	}

	searchResp, err := client.LexicalSearch(searchCtx, searchReq)
	if err != nil {
		t.Fatalf("LexicalSearch failed: %v", err)
	}
	if !searchResp.Success {
		t.Errorf("LexicalSearch was not successful: %s", searchResp.Message)
	}
	if len(searchResp.Results) == 0 {
		t.Errorf("LexicalSearch returned no results")
	}

	// Delete
	delCtx, delCancel := context.WithTimeout(context.Background(), time.Second)
	defer delCancel()

	delReq := &pb.DeleteRequest{
		IndexPath:  testIndexPath,
		DeleteCopy: true,
		DataUris:   []string{filepath.Join(testDir, "test_file1.txt")},
	}

	delResp, err := client.Delete(delCtx, delReq)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !delResp.Success {
		t.Errorf("Delete was not successful: %s", delResp.Message)
	}
}
