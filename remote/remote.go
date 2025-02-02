package remote

import (
	"fmt"
	"log/slog"

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
	config := cfg.KeysHash()
	slog.Debug(fmt.Sprintf("+%v", config))
	url, _ := cfg.GetKey("url")
	s := NewS3Remote(url.String(), config)
	return s, nil
}
