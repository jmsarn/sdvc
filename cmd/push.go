/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/jmsarn/sdvc/storage"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push file",
	Args:  cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: toggleDebug,
	RunE: func(cmd *cobra.Command, args []string) error {
		return pushFile(args[0])
	},
}

func pushFile(path string) error {
	cfg, err := readConfig(false)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
	}
	sec, _ := cfg.File.GetSection("core")
	key, err := sec.GetKey("remote")
	if err != nil {
		return errors.New("No default remote specified")
	}
	remote, err := cfg.File.GetSection(fmt.Sprintf(`remote "%s"`, key))
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error reading configuration for remote %s\n", key),
		)
	}
	remoteUrl, _ := remote.GetKey("url")

	// Update remote configuration with values from config.local
	cfgLocal, err := readConfig(true)
	if err == nil {
		remoteLocal, err := cfgLocal.File.GetSection(
			fmt.Sprintf(`remote "%s"`, key),
		)
		if err == nil {
			for k, v := range remoteLocal.KeysHash() {
				if remote.HasKey(k) {
					k_, _ := remote.GetKey(k)
					k_.SetValue(v)
				} else {
					remote.NewKey(k, v)
				}
			}
		}
	}

	remoteStorage := storage.NewStorage(remote)
	ptr, err := getPointerFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading SDVC file %s: %s", ptr.Path, err))
	}
	hash, err := fileHash(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error calculating hash for %s: %s", path, err))
	}
	if hash != ptr.SHA256 {
		return errors.New(
			fmt.Sprintf(`SHA256 of pointer and file don't match. Use sdvc add to stage latest changes to %s`, path),
		)
	}
	remotePath, _ := url.JoinPath(remoteUrl.String(), rootRelativePath(path))
	result, err := remoteStorage.Upload(
		storage.StorageObject{
			LocalPath:  path,
			RemotePath: remotePath,
			SHA256:     ptr.SHA256,
		},
	)
	if result == nil {
		fmt.Println("File already exists in remote, skipping upload")
		return nil
	}
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error uploading %s to %s: %s\n", path, remotePath, err),
		)
	}
	ptr.Cloud[key.String()] = CloudInfo{
		ETag:      result.ETag,
		VersionID: result.Version,
	}
	if err = writePointerFile(ptr, path); err != nil {
		return errors.New(
			fmt.Sprintf("Error reading writing pointer file %s: %s", path, err),
		)
	}
	cache, err := readCache()
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading cache file: %s", err))
	}
	cache.Files[rootRelativePath(path)] = ptr.SHA256
	if err = cache.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error updating cache: %s", err))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
