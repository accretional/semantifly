package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Main is run before every test to set up the testing folder & semantifly
func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("Failed to change directory: %v\n", err)
		os.Exit(1)
	}

	semantifly_dir := os.Getenv("HOME") + "/opt/semantifly"
	cmd := exec.Command("bash", "setup.sh")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Setup for semantifly failed: %v\nStderr: %s\n", err, stderr.String())
		os.Exit(1)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", oldPath+":"+semantifly_dir)

	// run tests
	code := m.Run()

	// clean up
	os.Setenv("PATH", oldPath)
	os.Remove("index.list")

	os.Exit(code)
}

func TestHelpFlag(t *testing.T) {
	expectedHelpString := "Use 'semantifly <subcommand> --help' for more information about a specific subcommand."

	t.Run("--help", func(t *testing.T) {
		if err := runAndCheckStdoutContains("--help", expectedHelpString, nil); err != nil {
			t.Errorf("Failed to execute --help flag: %v", err)
		}
	})

	t.Run("-h", func(t *testing.T) {
		if err := runAndCheckStdoutContains("-h", expectedHelpString, nil); err != nil {
			t.Errorf("Failed to execute -h flag: %v", err)
		}
	})
}

func TestAddSubcommand(t *testing.T) {
	tempFile, err := createTempFile("This is a test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile)

	t.Run("Help", func(t *testing.T) {
		if err := runAndCheckStderrContains("add", "Usage of add:", []string{"--help"}); err != nil {
			t.Errorf("Failed to execute 'add --help': %v", err)
		}
	})

	t.Run("Non-existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "file does not exist", []string{"nofile"}); err != nil {
			t.Errorf("Failed to execute 'add' with non-existing file: %v", err)
		}
	})

	t.Run("Bad flag", func(t *testing.T) {
		if err := runAndCheckStderrContains("add", "flag provided but not defined: -badflag", []string{"--badflag"}); err != nil {
			t.Errorf("Failed to execute 'add' with bad flag: %v", err)
		}
	})

	t.Run("Existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "added successfully", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'add' with existing file: %v", err)
		}
	})
}

func TestGetSubcommand(t *testing.T) {
	tempFile, err := createTempFile("This is a test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile)

	if err := runAndCheckStdoutContains("add", "added successfully", []string{tempFile}); err != nil {
		t.Fatalf("Failed to add file to index: %v", err)
	}

	t.Run("Get existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("get", "This is a test file", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'get' for existing file: %v", err)
		}
	})

	t.Run("Get after delete", func(t *testing.T) {
		runAndCheckStdoutContains("delete", "deleted successfully", []string{tempFile})
		if err := runAndCheckStdoutContains("get", "empty index file", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'get' after delete: %v", err)
		}
	})
}

func TestUpdateSubcommand(t *testing.T) {
	tempFile, err := createTempFile("This is a test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile)

	updatedTempFile, err := createTempFile("This is an updated test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create updated temporary file: %v", err)
	}
	defer os.Remove(updatedTempFile)

	runAndCheckStdoutContains("add", "added successfully", []string{tempFile})

	t.Run("Update without URI", func(t *testing.T) {
		if err := runAndCheckStdoutContains("update", "Update subcommand requires two input args", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'update' without URI: %v", err)
		}
	})

	t.Run("Update with new file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("update", "updated successfully", []string{tempFile, updatedTempFile}); err != nil {
			t.Errorf("Failed to execute 'update' with new file: %v", err)
		}
	})
}

func TestDeleteSubcommand(t *testing.T) {
	tempFile, err := createTempFile("This is a test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile)

	runAndCheckStdoutContains("add", "added successfully", []string{tempFile})

	t.Run("Delete existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "deleted successfully", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'delete' for existing file: %v", err)
		}
	})

	t.Run("Delete from empty index", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "empty index file", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'delete' on empty index: %v", err)
		}
	})
}

func TestWebpageOperations(t *testing.T) {
	testWebpageURI := "http://echo.jsontest.com/title/lorem/content/ipsum"
	updatedWebpageURI := "http://echo.jsontest.com/title/foo/content/bar"

	t.Run("Add webpage", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "added successfully", []string{testWebpageURI}); err != nil {
			t.Errorf("Failed to execute 'add' for webpage: %v", err)
		}
	})

	t.Run("Get webpage", func(t *testing.T) {
		webpageContent := getWebpageContent(t, testWebpageURI)
		if err := runAndCheckStdoutContains("get", string(webpageContent), []string{testWebpageURI}); err != nil {
			t.Errorf("Failed to execute 'get' for webpage: %v", err)
		}
	})

	// to do: after we refactor name & uri to be different, need to change the tests after update to get / delete new name, not old name
	t.Run("Update webpage", func(t *testing.T) {
		if err := runAndCheckStdoutContains("update", "updated successfully", []string{testWebpageURI, updatedWebpageURI}); err != nil {
			t.Errorf("Failed to execute 'update' for webpage: %v", err)
		}
	})

	t.Run("Get webpage", func(t *testing.T) {
		webpageContent := getWebpageContent(t, updatedWebpageURI)
		if err := runAndCheckStdoutContains("get", string(webpageContent), []string{testWebpageURI}); err != nil {
			t.Errorf("Failed to execute 'get' for webpage: %v", err)
		}
	})

	t.Run("Delete webpage", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "deleted successfully", []string{testWebpageURI}); err != nil {
			t.Errorf("Failed to execute 'delete' for webpage: %v", err)
		}
	})
}

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

func runAndCheckStderrContains(subcommand string, wantedStderrSubstr string, args []string) error {
	allArgs := append([]string{subcommand}, args...)
	cmd := exec.Command("semantifly", allArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 2 {
			return fmt.Errorf("Command execution failed: %v\nStderr: %s", err, stderr.String())
		}
	}
	stderrOutput := stderr.String()

	if !strings.Contains(stderrOutput, wantedStderrSubstr) {
		return fmt.Errorf("Expected stderr to contain \"%s\". Stderr obtained \"%s\"", wantedStderrSubstr, stderrOutput)
	}

	return nil
}

func createTempFile(content string) (string, error) {
	tempFile, err := os.CreateTemp("", "semantifly_test_*.txt")
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary file: %v", err)
	}

	if _, err := tempFile.WriteString(content); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	return tempFile.Name(), nil
}

func getWebpageContent(t *testing.T, url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to fetch web page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("web page returned non-OK status: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read web page content: %v", err)
	}

	return content
}
