package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmsarn/sdvc/remote"
)

type mockRemote struct{}

func (r *mockRemote) CheckLocalObject(obj remote.FileObject) (bool, error) {
	return true, nil
}

func (r *mockRemote) Download(obj remote.FileObject) error {
	return nil
}

func (r *mockRemote) GetSHA256(obj remote.FileObject) (string, error) {
	return "sha256", nil
}

func (r *mockRemote) Upload(obj remote.FileObject) (*remote.UploadResult, error) {
	return &remote.UploadResult{
		ETag:    "etag",
		Path:    obj.LocalPath,
		SHA256:  "sha256",
		Version: "test",
	}, nil
}

func Test_ExecutePushCommand(t *testing.T) {
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
}
