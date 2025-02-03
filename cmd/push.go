/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmsarn/sdvc/remote"
	"github.com/jmsarn/sdvc/utils"
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
	PreRun: toggleVerbose,
	RunE: func(cmd *cobra.Command, args []string) error {
		remoteName, _ := cmd.Flags().GetString("remote")
		remote, err := remote.NewRemote(remoteName)
		if err != nil {
			return err
		}
		return pushFile(args[0], remoteName, remote)
	},
}

func pushFile(path, remoteName string, remoteStorage remote.Remote) error {
	path = strings.TrimSuffix(path, ".sdvc")
	ptr, err := getPointerFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading SDVC file %s: %s", ptr.Path, err))
	}
	hash, err := utils.FileSHA256(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error calculating hash for %s: %s", path, err))
	}
	if hash != ptr.SHA256 {
		return errors.New(
			fmt.Sprintf(`SHA256 of pointer and file don't match. Use sdvc add to stage latest changes to %s`, path),
		)
	}
	result, err := remoteStorage.Upload(
		remote.FileObject{
			LocalPath: utils.RootRelativePath(path),
			SHA256:    ptr.SHA256,
		},
	)
	if err != nil {
		return err
	}
	if result == nil {
		fmt.Println("File already exists in remote, skipping upload")
		return nil
	}
	ptr.Cloud[remoteName] = CloudInfo{
		ETag:      result.ETag,
		VersionID: result.Version,
	}
	if err = writePointerFile(ptr, path); err != nil {
		return errors.New(
			fmt.Sprintf("Error reading writing pointer file %s: %s", path, err),
		)
	}
	cache, err := utils.ReadCache()
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading cache file: %s", err))
	}
	cache.Files[utils.RootRelativePath(path)] = ptr.SHA256
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
	pushCmd.Flags().String("remote", "default", "Remote destination of the file")
}
