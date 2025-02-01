package storage

import (
	"gopkg.in/ini.v1"
)

type StorageObject struct {
	LocalPath  string
	RemotePath string
	SHA256     string
	Version    string
}

type UploadResult struct {
	ETag    string
	Path    string
	SHA256  string
	Version string
}

type Storage interface {
	CheckLocalObject(obj StorageObject) (bool, error)
	Download(obj StorageObject) error
	GetSHA256(obj StorageObject) (string, error)
	Upload(obj StorageObject) (*UploadResult, error)
}

func NewStorage(cfg *ini.Section) Storage {
	url, _ := cfg.GetKey("url")
	s := NewS3Storage(url.String(), cfg.KeysHash())
	return s
}
