package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type Cache struct {
	Files map[string]string
}

func (c *Cache) Save() error {
	configPath, err := ProjectConfigPath()
	if err != nil {
		return err
	}
	path := filepath.Join(filepath.Dir(configPath), ".cache")
	data, err := yaml.Marshal(&c.Files)
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil
}

type Config struct {
	Path string
	File *ini.File
}

func (c *Config) Save() error {
	err := c.File.SaveTo(c.Path)
	return err
}

func Exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func ProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Traverse up the directory tree looking for .git directory
	for {
		gitDir := filepath.Join(currentDir, ".git")
		if Exists(gitDir) {
			return currentDir, nil
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", os.ErrNotExist
		}
		currentDir = parentDir
	}
}

func RootRelativePath(path string) string {
	pRoot, _ := ProjectRoot()
	p, _ := filepath.Rel(pRoot, filepath.Join(pRoot, path))
	return p
}

func ProjectConfigPath() (string, error) {
	gitRoot, err := ProjectRoot()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(gitRoot, ".sdvc", "config")
	return configPath, nil
}

func ReadConfig(local bool) (*Config, error) {
	configPath, err := ProjectConfigPath()
	if err != nil {
		return nil, err
	}
	var cfg *ini.File
	if local {
		configPath = fmt.Sprint(configPath, ".local")
		if !Exists(configPath) {
			cfg = ini.Empty()
			if err = cfg.SaveTo(configPath); err != nil {
				return nil, err
			}
		}
	}
	cfg, err = ini.Load(configPath)
	if err != nil {
		return nil, err
	}
	return &Config{configPath, cfg}, nil
}

func ReadRemoteConfig(remoteName string) (*ini.Section, error) {
	cfg, err := ReadConfig(false)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
	}
	if remoteName == "default" {
		sec, _ := cfg.File.GetSection("core")
		key, err := sec.GetKey("remote")
		if err != nil {
			return nil, errors.New("No default remote specified")
		}
		remoteName = key.String()
	}
	r, err := cfg.File.GetSection(fmt.Sprintf(`remote "%s"`, remoteName))
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("Error reading configuration for remote %s\n", remoteName),
		)
	}
	// Update remote configuration with values from config.local
	cfgLocal, err := ReadConfig(true)
	if err == nil {
		rLocal, err := cfgLocal.File.GetSection(
			fmt.Sprintf(`remote "%s"`, remoteName),
		)
		if err == nil {
			for k, v := range rLocal.KeysHash() {
				if r.HasKey(k) {
					k_, _ := r.GetKey(k)
					k_.SetValue(v)
				} else {
					r.NewKey(k, v)
				}
			}
		}
	}
	return r, nil
}

func ReadCache() (*Cache, error) {
	configPath, err := ProjectConfigPath()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(filepath.Dir(configPath), ".cache")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files := map[string]string{}
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&files); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}
	return &Cache{Files: files}, nil
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash, nil
}
