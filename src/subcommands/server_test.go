package subcommands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const testDir = "./test_semantifly"

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
	startTestServer()

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
	}

	return client, cleanup
}

func startTestServer() {
	go executeStartServer([]string{"--index-path", testDir})

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)
}

func TestServerCommands(t *testing.T) {
	// Start server
	cleanupTestEnv := setupTestEnvironment(t)
	defer cleanupTestEnv()

	client, cleanupClient := setupServerAndClient(t)
	defer cleanupClient()

	// Add
	addCtx, addCancel := context.WithTimeout(context.Background(), time.Second)
	defer addCancel()

	testFileData1 := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        filepath.Join(testDir, "test_file1.txt"),
	}

	addArgs := &pb.AddRequest{
		AddedMetadata: testFileData1,
		MakeCopy:      true,
	}

	addResp, err := client.Add(addCtx, addArgs)

	if err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}
	if addResp.ErrorMessage != "" {
		t.Fatalf("Failed to add test file: %v", err)
	}

	// Add again
	addCtx2, addCancel2 := context.WithTimeout(context.Background(), time.Second)
	defer addCancel2()

	testFileData2 := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        filepath.Join(testDir, "test_file1.txt"),
	}

	addArgs2 := &pb.AddRequest{
		AddedMetadata: testFileData2,
		MakeCopy:      true,
	}

	_, err = client.Add(addCtx2, addArgs2)

	if err.Error() == "" {
		t.Fatalf("Failed to return error message when expected to")
	}

	expectedErrorMsg := "make_copy:true has already been added. Skipping without refresh."
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Fatalf("Returned error message %v when expected %v", err.Error(), expectedErrorMsg)
	}

	// Get
	getCtx, getCancel := context.WithTimeout(context.Background(), time.Second)
	defer getCancel()

	getReq := &pb.GetRequest{
		Name: filepath.Join(testDir, "test_file1.txt"),
	}

	getResp, err := client.Get(getCtx, getReq)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	expectedContent := "This is a test file for Semantifly"
	if *getResp.Content == "" {
		t.Errorf("Get returned empty content")
	} else if *getResp.Content != expectedContent {
		t.Errorf("Get returned content %v when expected %v", *getResp.Content, expectedContent)
	}

	// Update
	updCtx, updCancel := context.WithTimeout(context.Background(), time.Second)
	defer updCancel()

	testUpdateFileData := &pb.ContentMetadata{
		DataType:   0,
		SourceType: 0,
		URI:        filepath.Join(testDir, "test_file2.txt"),
	}

	updReq := &pb.UpdateRequest{
		Name:            filepath.Join(testDir, "test_file1.txt"),
		UpdatedMetadata: testUpdateFileData,
		UpdateCopy:      true,
	}

	updResp, err := client.Update(updCtx, updReq)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updResp.ErrorMessage != "" {
		t.Fatalf("Failed to add test file: %v", err)
	}

	// Search
	searchCtx, searchCancel := context.WithTimeout(context.Background(), time.Second)
	defer searchCancel()

	searchReq := &pb.LexicalSearchRequest{
		SearchTerm: "test",
		TopN:       5,
	}

	searchResp, err := client.LexicalSearch(searchCtx, searchReq)
	if err != nil {
		t.Fatalf("LexicalSearch failed: %v", err)
	}
	if len(searchResp.Results) == 0 {
		t.Errorf("LexicalSearch returned no results")
	}

	expectedResult := &pb.LexicalSearchResult{
		Name:        "test_semantifly/test_file1.txt",
		Occurrences: 2,
	}

	result := searchResp.Results[0]

	if (result.Name != expectedResult.Name) || (result.Occurrences != expectedResult.Occurrences) {
		t.Fatalf("Lexical Search returned %v, expected %v", result, expectedResult)
	}

	// Delete unnamed file check error
	badDelCtx, badDelCancel := context.WithTimeout(context.Background(), time.Second)
	defer badDelCancel()

	badDelReq := &pb.DeleteRequest{
		DeleteCopy: true,
		Names:      []string{filepath.Join(testDir, "bad_file.txt")},
	}

	expectedDelError := "Entry test_semantifly/bad_file.txt not found in index file"

	badDelResp, err := client.Delete(badDelCtx, badDelReq)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if badDelResp.ErrorMessage == "" {
		t.Fatalf("Failed to return error message when error message expected.")
	} else if !strings.Contains(badDelResp.ErrorMessage, expectedDelError) {
		t.Fatalf("Expected message %v to be inside error, got %v", expectedDelError, badDelResp.ErrorMessage)
	}

	// Delete
	delCtx, delCancel := context.WithTimeout(context.Background(), time.Second)
	defer delCancel()

	delReq := &pb.DeleteRequest{
		DeleteCopy: true,
		Names:      []string{filepath.Join(testDir, "test_file1.txt")},
	}

	delResp, err := client.Delete(delCtx, delReq)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if delResp.ErrorMessage != "" {
		t.Fatalf("Failed to add test file: %v", err)
	}
}
