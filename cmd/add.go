/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type CloudInfo struct {
	ETag      string `yaml:"etag"`
	VersionID string `yaml:"version_id"`
}

type FilePointer struct {
	SHA256 string               `yaml:"sha256"`
	Size   int64                `yaml:"size"`
	Hash   string               `yaml:"hash"`
	Path   string               `yaml:"path"`
	Cloud  map[string]CloudInfo `yaml:"cloud"`
}

type FilePointerOutput struct {
	Outs []*FilePointer `yaml:"outs"`
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:    "add file",
	Args:   cobra.ExactArgs(1),
	Short:  "Track data files or directories with SDVC",
	Long:   "Track data files or directories with SDVC",
	PreRun: toggleDebug,
	RunE: func(cmd *cobra.Command, args []string) error {
		return addFile(args[0])
	},
}

func addFile(path string) error {
	ptrFile, err := getPointerFile(path)
	if err != nil {
		ptrPath := fmt.Sprintf("%s.sdvc", path)
		return errors.New(
			fmt.Sprintf("Error reading SDVC file %s: %s\n", ptrPath, err),
		)
	}
	fileInfo, err := getFile(path)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error getting file information for %s: %s\n", path, err),
		)
	}
	if ptrFile == nil || ptrFile.SHA256 != fileInfo.SHA256 {
		err = writePointerFile(fileInfo, path)
		if err != nil {
			ptrPath := fmt.Sprintf("%s.sdvc", path)
			return errors.New(
				fmt.Sprintf("Error writing SDVC file %s: %s\n", ptrPath, err),
			)
		}
		cache, err := readCache()
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading cache file: %s", err))
		}
		cache.Files[rootRelativePath(path)] = fileInfo.SHA256
		if err = cache.Save(); err != nil {
			return errors.New(fmt.Sprintf("Error updating cache: %s", err))
		}
		return nil
	}
	fmt.Println("Existing SDVC file matches current file, nothing to add")
	return nil

}

func writePointerFile(ptr *FilePointer, path string) error {
	if !strings.HasSuffix(path, ".sdvc") {
		path = fmt.Sprintf("%s.sdvc", path)
	}
	data, err := yaml.Marshal(&FilePointerOutput{Outs: []*FilePointer{ptr}})
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil
}

func getPointerFile(path string) (*FilePointer, error) {
	if !strings.HasSuffix(path, ".sdvc") {
		path = fmt.Sprintf("%s.sdvc", path)
	}
	if exists(path) {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		var ptrOut FilePointerOutput
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&ptrOut); err != nil {
			return nil, err
		}
		return ptrOut.Outs[0], nil
	}
	return nil, nil
}

func getFile(path string) (*FilePointer, error) {

	hash, err := fileHash(path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	ptr := &FilePointer{
		SHA256: hash,
		Size:   info.Size(),
		Hash:   "sha256",
		Path:   filepath.Base(path),
	}
	return ptr, nil
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
