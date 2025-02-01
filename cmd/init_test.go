package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func cleanUpConfig() {
	configPath, _ := getProjectConfigPath()
	configDir := filepath.Dir(configPath)
	os.RemoveAll(configDir)
}

func compareFiles(pathA, pathB string) error {
	contentA, err := os.ReadFile(pathA)
	if err != nil {
		return err
	}
	contentB, err := os.ReadFile(pathB)
	if err != nil {
		return err
	}
	if !bytes.Equal(contentA, contentB) {
		output := new(bytes.Buffer)
		output.WriteString(fmt.Sprintf("%q and %q are not equal!\n\n", pathA, pathB))

		diffPath, err := exec.LookPath("diff")
		if err != nil {
			// Don't execute diff if it can't be found.
			return nil
		}
		diffCmd := exec.Command(diffPath, "-u", "--strip-trailing-cr", pathA, pathB)
		diffCmd.Stdout = output
		diffCmd.Stderr = output

		output.WriteString("$ diff -u " + pathA + " " + pathB + "\n")
		if err := diffCmd.Run(); err != nil {
			output.WriteString("\n" + err.Error())
		}
		return errors.New(output.String())
	}
	return nil
}

func TestExecuteAddCommand(t *testing.T) {
	cleanUpConfig()
	defer cleanUpConfig()

	err := initializeProject("s3://test-bucket", false)
	if err != nil {
		t.Fatal(err)
	}

	configPath, _ := getProjectConfigPath()
	goldenFile := filepath.Join("test-data", "config.golden")
	err = compareFiles(configPath, goldenFile)
	if err != nil {
		t.Fatal(err)
	}
}
