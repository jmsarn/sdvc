/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/jmsarn/sdvc/remote"
	"github.com/jmsarn/sdvc/utils"
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
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
		return pullFile(args[0], remoteName, remote)
	},
}

func pullFile(path, remoteName string, remoteStorage remote.Remote) error {
	ptr, err := getPointerFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading SDVC file %s: %s", ptr.Path, err))
	}
	hash, err := utils.FileSHA256(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error calculating hash for %s: %s", path, err))
	}
	cache, err := utils.ReadCache()
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading cache file: %s", err))
	}
	if cache.Files[utils.RootRelativePath(path)] != hash {
		return errors.New(
			fmt.Sprintf(
				`SHA256 of cache and file don't match. Did you modify %s without pushing?`,
				path,
			),
		)
	}
	if err := remoteStorage.Download(
		remote.FileObject{
			LocalPath: utils.RootRelativePath(path),
			SHA256:    ptr.SHA256,
			Version:   ptr.Cloud[remoteName].VersionID,
		},
	); err != nil {
		return err
	}
	cache.Files[utils.RootRelativePath(path)] = ptr.SHA256
	if err = cache.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error updating cache: %s", err))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
