package tests

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"accretional.com/semantifly/subcommands"

	pb "accretional.com/semantifly/proto/accretional.com/semantifly/proto"
)

// captures and returns output from console
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	defer r.Close()

	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	f()
	w.Close()

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestGet(t *testing.T) {
	// set up temporary directory
	indexDir, err := os.MkdirTemp("", "test_get")

	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(indexDir)

	srcContent := "Test Content"
	srcFile := createTempFile(t, indexDir, srcContent)

	addArgs := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   true,
		DataURIs:   []string{srcFile.Name()},
	}

	subcommands.Add(addArgs)

	srcContent2 := "Other Content"
	srcFile2 := createTempFile(t, indexDir, srcContent2)

	addArgs2 := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   true,
		DataURIs:   []string{srcFile2.Name()},
	}

	subcommands.Add(addArgs2)

	getArgs := subcommands.GetArgs{
		IndexPath: indexDir,
		FileName:  srcFile.Name(),
	}

	output := captureOutput(func() {
		err = subcommands.Get(getArgs)
	})

	// validate the output and error
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	// add new line because of fmt.PrintLn()
	expectedOutput := srcContent + "\n"

	if output != expectedOutput {
		t.Fatalf("Expected content %q, got %q", expectedOutput, output)
	}
}

func TestGetError(t *testing.T) {
	indexDir, _ := os.MkdirTemp("", "test_get")

	nonExistentFile := "non_existent_file.txt"
	getArgsNonExistent := subcommands.GetArgs{
		IndexPath: indexDir,
		FileName:  nonExistentFile,
	}

	srcContent := "Existing Content"
	srcFile := createTempFile(t, indexDir, srcContent)

	args := subcommands.AddArgs{
		IndexPath:  indexDir,
		DataType:   pb.DataType_TEXT,
		SourceType: pb.SourceType_LOCAL_FILE,
		MakeCopy:   true,
		DataURIs:   []string{srcFile.Name()},
	}

	subcommands.Add(args)

	err := subcommands.Get(getArgsNonExistent)

	// Validate the error
	if err == nil {
		t.Fatalf("Expected an error for non-existent file, but got nil")
	}

	expectedErrMsg := fmt.Sprintf("file '%s' not found in the index", nonExistentFile)
	if err.Error() != expectedErrMsg {
		t.Fatalf("Expected error message %q, but got %q", expectedErrMsg, err.Error())
	}
}
