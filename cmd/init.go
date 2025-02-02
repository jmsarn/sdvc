/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmsarn/sdvc/utils"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:    "init",
	Short:  "Initialize SDVC in the current project. Expects project to be a Git repository",
	Long:   "Initialize SDVC in the current ptoject. Expects project to be a Git repository",
	PreRun: toggleVerbose,
	RunE: func(cmd *cobra.Command, args []string) error {
		remote, _ := cmd.Flags().GetString("remote")
		force, _ := cmd.Flags().GetBool("force")
		return initializeProject(remote, force)
	},
}

func initializeProject(remote string, force bool) error {
	configPath, err := utils.ProjectConfigPath()
	if err != nil {
		return errors.New("Cannot find project root. Are you in a Git repository?")
	}
	configDir := filepath.Dir(configPath)
	rootDir := filepath.Dir(configDir)
	dirInfo, err := os.Stat(rootDir)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot stat %s: %s\n", rootDir, err))
	}
	if utils.Exists(configDir) {
		if force {
			os.RemoveAll(configDir)
		} else {
			fmt.Println(`Warning: existing config file found
Use --force to overwrite existing configuration
					`)
			return nil
		}
	}
	os.Mkdir(configDir, dirInfo.Mode())
	cfg := ini.Empty()
	core, _ := cfg.NewSection("core")
	if remote != "" {
		core.NewKey("remote", "default")
		sec, _ := cfg.NewSection(`remote "default"`)
		sec.NewKey("url", remote)
	}
	err = cfg.SaveTo(filepath.Join(configDir, "config"))
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error creating config file %s: %s", configPath, err),
		)
	}
	ignore, _ := os.Create(filepath.Join(configDir, ".gitignore"))
	defer ignore.Close()
	ignore.WriteString("config.local\n")
	ignore.Sync()

	cache, _ := os.OpenFile(filepath.Join(configDir, ".cache"), os.O_RDONLY|os.O_CREATE, 0644)
	defer cache.Close()
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	initCmd.Flags().String("remote", "", "URL for default remote")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration")
}
