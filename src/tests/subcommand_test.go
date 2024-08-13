package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// runAndAssertSubcommand executes a subcommand of the "semantifly" command with the provided arguments
// and asserts that the output contains the specified assertion statement.
//
// Parameters:
//   - subcommand (string): The subcommand to execute.
//   - assertStatement (string): The expected substring to be present in the command output.
//   - args ([]string): Additional arguments to pass to the subcommand.
func runAndAssertSubcommand(subcommand string, assertStatement string, args []string) error {
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

	// Assertion for command output
	if !strings.Contains(output, assertStatement) {
		return fmt.Errorf("Expected output to contain %s, but got: %s", assertStatement, output)
	}

	return nil
}

func TestCommandRun(t *testing.T) {
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	semantifly_dir := os.Getenv("HOME") + "/opt/semantifly"
	os.Setenv("PATH", oldPath+":"+semantifly_dir)

	tempFile, err := os.CreateTemp("", "semantifly_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "This is a test file for semantifly add command."
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	args := []string{tempFile.Name()}

	// Testing Add subcommand
	if err := runAndAssertSubcommand("add", "Added file successfully", args); err != nil {
		t.Errorf("Failed to execute 'add' subcommand: %v", err)
	}

	// Testing Get subcommand
	if err := runAndAssertSubcommand("get", testContent, args); err != nil {
		t.Errorf("Failed to execute 'add' subcommand: %v", err)
	}

	// Testing Delete subcommand
	if err := runAndAssertSubcommand("delete", "Deleted entry from index", args); err != nil {
		t.Errorf("Failed to execute 'delete' subcommand: %v", err)
	}
}
