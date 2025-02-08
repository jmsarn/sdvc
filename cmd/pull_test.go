package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExecutePullCommand(t *testing.T) {
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
	defer os.RemoveAll(filepath.Join(".", "test"))

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

	remoteStorage := mockRemote{}
	err = pushFile(testFile, "default", &remoteStorage)
	if err != nil {
		t.Fatal(err)
	}

	err = pullFile(
		fmt.Sprintf("%s.sdvc", testFile),
		"default",
		&remoteStorage,
	)
	if err != nil {
		t.Fatal(err)
	}
}
