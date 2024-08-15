package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runAndCheckStdoutContains(subcommand string, wantedStdoutSubstr string, args []string) error {
	allArgs := append([]string{subcommand}, args...)
	cmd := exec.Command("semantifly", allArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Command execution failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()

	if !strings.Contains(output, wantedStdoutSubstr) {
		return fmt.Errorf("Expected output to contain \"%s\". Output obtained \"%s\"", wantedStdoutSubstr, output)
	}

	return nil
}

func TestCommandRun(t *testing.T) {

	// Setup
	err := os.Chdir("..")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	semantifly_dir := os.Getenv("HOME") + "/opt/semantifly"
	cmd := exec.Command("bash", "setup.sh")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("Setup for semantifly failed: %v\nStderr: %s", err, stderr.String())
	}

	// Adding semantifly to PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	os.Setenv("PATH", oldPath+":"+semantifly_dir)

	tempFile, err := os.CreateTemp("", "semantifly_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "This is a test file for semantifly subcommands."
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Making a second file to test the update command 
	updatedTempFile, err := os.CreateTemp("", "semantifly_test_updated_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(updatedTempFile.Name())

	updatedContent := "This is an updated test file for semantifly subcommands."
	if _, err := updatedTempFile.WriteString(updatedContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Testing Add subcommand for a non-existing file
	args := []string{"nofile"}
	if err := runAndCheckStdoutContains("add", "file does not exist", args); err != nil {
		t.Errorf("Failed to execute 'add' subcommand: %v", err)
	}

	// Testing Add subcommand for an existing file
	args = []string{tempFile.Name()}
	if err := runAndCheckStdoutContains("add", "added successfully", args); err != nil {
		t.Errorf("Failed to execute 'add' subcommand: %v", err)
	}
	defer os.Remove("index.list")

	// Testing Get subcommand
	if err := runAndCheckStdoutContains("get", testContent, args); err != nil {
		t.Errorf("Failed to execute 'get' subcommand: %v", err)
	}

	// Testing Update subcommand without passing in the updated URI
	if err := runAndCheckStdoutContains("update", "Update subcommand requires two input args", args); err != nil {
		t.Errorf("Failed to execute 'delete' subcommand: %v", err)
	}

	// Testing Update subcommand
	updateArgs := []string{tempFile.Name(), updatedTempFile.Name()}
	if err := runAndCheckStdoutContains("update", "updated successfully", updateArgs); err != nil {
		t.Errorf("Failed to execute 'delete' subcommand: %v", err)
	}

	// Testing Delete subcommand
	if err := runAndCheckStdoutContains("delete", "deleted successfully", args); err != nil {
		t.Errorf("Failed to execute 'delete' subcommand: %v", err)
	}

	// Testing Get command after deleting the entry
	if err := runAndCheckStdoutContains("get", "empty index file", args); err != nil {
		t.Errorf("Failed to execute 'get' subcommand: %v", err)
	}

	// Testing Delete subcommand after deleting the entry
	if err := runAndCheckStdoutContains("delete", "empty index file", args); err != nil {
		t.Errorf("Failed to execute 'delete' subcommand: %v", err)
	}

}
