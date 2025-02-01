package remote

import (
	"errors"
	"fmt"
	"github.com/jmsarn/sdvc/utils"
)

type FileObject struct {
	LocalPath string
	SHA256    string
	Version   string
}

type UploadResult struct {
	ETag    string
	Path    string
	SHA256  string
	Version string
}

type Remote interface {
	CheckLocalObject(obj FileObject) (bool, error)
	Download(obj FileObject) error
	GetSHA256(obj FileObject) (string, error)
	Upload(obj FileObject) (*UploadResult, error)
}

func NewRemote(remoteName string) (Remote, error) {
	cfg, err := utils.ReadRemoteConfig(remoteName)
	if err != nil {
		return nil, err
	}
	url, _ := cfg.GetKey("url")
	if isValidS3URI(url.String()) {
		s := NewS3Remote(url.String(), cfg.KeysHash())
		return s, nil
	}
	return nil, errors.New(fmt.Sprintf("Invalid remote URL: %s check your config", url))
}
