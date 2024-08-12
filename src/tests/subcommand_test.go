package tests

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestCommandRun_Add(t *testing.T) {
	oldPath := os.Getenv("PATH")
	homePath := os.Getenv("HOME")
	os.Setenv("PATH", oldPath+":"+homePath+"/opt/semantifly")

	defer os.Setenv("PATH", oldPath)

	cmd := exec.Command("semantifly", "add", homePath+"/semantifly/README.md")

	op, err := cmd.Output()
	fmt.Printf("Output: %s\n", op)

	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
}
