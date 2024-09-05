package subcommands

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testIndexPath string

// Main is run before every test to set up the testing folder & semantifly
func TestMain(m *testing.M) {
	err := os.Chdir("../..")
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

	testDir, err := os.MkdirTemp("", "semantifly-test")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	testIndexPath = testDir

	// run tests
	code := m.Run()

	// clean up
	os.Setenv("PATH", oldPath)

	testIndex := filepath.Join(testIndexPath, "index.list")
	os.Remove(testIndex)

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
	testIndex := filepath.Join(testIndexPath, "index.list")
	newTestIndexPath := filepath.Join(testIndexPath, "test-in-test")
	os.Remove(testIndex)

	t.Run("Help", func(t *testing.T) {
		if err := runAndCheckStderrContains("add", "Usage of add:", []string{"--help"}); err != nil {
			t.Errorf("Failed to execute 'add --help': %v", err)
		}
	})

	t.Run("Non-existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "file does not exist", []string{"nofile", "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'add' with non-existing file: %v", err)
		}
	})

	t.Run("Bad flag", func(t *testing.T) {
		if err := runAndCheckStderrContains("add", "flag provided but not defined: -badflag", []string{"--badflag"}); err != nil {
			t.Errorf("Failed to execute 'add' with bad flag: %v", err)
		}
	})

	t.Run("Existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "", []string{tempFile, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'add' with existing file: %v", err)
		}
	})

	t.Run("Creates new directory for index", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "", []string{tempFile, "--index-path", newTestIndexPath}); err != nil {
			t.Errorf("Failed to execute 'add' with existing file: %v", err)
		}

		if _, err := os.Stat(newTestIndexPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to be created, but it doesn't exist", newTestIndexPath)
		}

		os.RemoveAll(newTestIndexPath)
	})
}

func TestGetSubcommand(t *testing.T) {
	tempFile, err := createTempFile("This is a test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile)
	testIndex := filepath.Join(testIndexPath, "index.list")
	badIndexPath := "bad/path"
	os.Remove(testIndex)

	if err := runAndCheckStdoutContains("add", "", []string{tempFile, "--index-path", testIndexPath}); err != nil {
		t.Fatalf("Failed to add file to index: %v", err)
	}

	t.Run("Get existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("get", "This is a test file", []string{tempFile, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'get' for existing file: %v", err)
		}
	})

	t.Run("Get after delete", func(t *testing.T) {
		runAndCheckStdoutContains("delete", "", []string{tempFile, "--index-path", testIndexPath})
		if err := runAndCheckStdoutContains("get", "empty index file", []string{tempFile, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'get' after delete: %v", err)
		}
	})

	t.Run("Get bad index file", func(t *testing.T) {
		runAndCheckStdoutContains("delete", "", []string{tempFile, "--index-path", testIndexPath})
		if err := runAndCheckStdoutContains("get", "no such file or directory", []string{tempFile, "--index-path", badIndexPath}); err != nil {
			t.Errorf("Failed to execute 'get' on bad index path: %v", err)
		}
	})
}

func TestUpdateSubcommand(t *testing.T) {
	tempFile, err := createTempFile("This is a test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile)
	testIndex := filepath.Join(testIndexPath, "index.list")
	os.Remove(testIndex)

	updatedTempFile, err := createTempFile("This is an updated test file for semantifly subcommands.")
	if err != nil {
		t.Fatalf("Failed to create updated temporary file: %v", err)
	}
	defer os.Remove(updatedTempFile)

	runAndCheckStdoutContains("add", "", []string{tempFile, "--index-path", testIndexPath})
	t.Run("Update without URI", func(t *testing.T) {
		if err := runAndCheckStdoutContains("update", "Update subcommand requires two input args", []string{tempFile}); err != nil {
			t.Errorf("Failed to execute 'update' without URI: %v", err)
		}
	})

	t.Run("Update with new file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("update", "", []string{tempFile, updatedTempFile, "--index-path", testIndexPath}); err != nil {
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
	testIndex := filepath.Join(testIndexPath, "index.list")
	badIndexPath := "bad/path"
	os.Remove(testIndex)

	runAndCheckStdoutContains("add", "", []string{tempFile, "--index-path", testIndexPath})

	t.Run("Delete bad index file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "no such file or directory", []string{tempFile, "--index-path", badIndexPath}); err != nil {
			t.Errorf("Failed to execute 'delete' on bad index path: %v", err)
		}
	})

	t.Run("Delete existing file", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "", []string{tempFile, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'delete' for existing file: %v", err)
		}
	})

	t.Run("Delete from empty index", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "empty index file", []string{tempFile, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'delete' on empty index: %v", err)
		}
	})
}

func TestWebpageOperations(t *testing.T) {
	testWebpageURI := "http://echo.jsontest.com/title/lorem/content/ipsum"
	updatedWebpageURI := "http://echo.jsontest.com/title/foo/content/bar"

	testIndex := filepath.Join(testIndexPath, "index.list")
	os.Remove(testIndex)
	t.Run("Add webpage", func(t *testing.T) {
		if err := runAndCheckStdoutContains("add", "", []string{testWebpageURI, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'add' for webpage: %v", err)
		}
	})

	t.Run("Get webpage", func(t *testing.T) {
		webpageContent := getWebpageContent(t, testWebpageURI)
		if err := runAndCheckStdoutContains("get", string(webpageContent), []string{testWebpageURI, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'get' for webpage: %v", err)
		}
	})

	// to do: after we refactor name & uri to be different, need to change the tests after update to get / delete new name, not old name
	t.Run("Update webpage", func(t *testing.T) {
		if err := runAndCheckStdoutContains("update", "", []string{testWebpageURI, updatedWebpageURI, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'update' for webpage: %v", err)
		}
	})

	t.Run("Get webpage", func(t *testing.T) {
		webpageContent := getWebpageContent(t, updatedWebpageURI)
		if err := runAndCheckStdoutContains("get", string(webpageContent), []string{testWebpageURI, "--index-path", testIndexPath}); err != nil {
			t.Errorf("Failed to execute 'get' for webpage: %v", err)
		}
	})

	t.Run("Delete webpage", func(t *testing.T) {
		if err := runAndCheckStdoutContains("delete", "", []string{testWebpageURI, "--index-path", testIndexPath}); err != nil {
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
