package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExecuteAddCommand(t *testing.T) {
	cleanUpConfig()
	defer cleanUpConfig()

	err := initializeProject("s3://test-bucket", false)
	if err != nil {
		t.Fatal(err)
	}

	testDataDir := filepath.Join(".", "test/data")
	err = os.MkdirAll(testDataDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(testDataDir)

	testFile := filepath.Join(testDataDir, "test.txt")
	f := []byte("this is a test data file")
	err = os.WriteFile(testFile, f, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = addFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	ptrFile := fmt.Sprintf("%s.sdvc", testFile)
	assert.FileExists(t, ptrFile, "Expected pointer file to exist")

	goldenFile := filepath.Join("test-data", "test.txt.sdvc.golden")
	err = compareFiles(ptrFile, goldenFile)
	if err != nil {
		t.Fatal(err)
	}
}
