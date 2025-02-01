/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/jmsarn/sdvc/utils"
	"github.com/spf13/cobra"
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:    "remote",
	Short:  "Setup and manage data remotes",
	Long:   `Setup and manage data remotes`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("remote called")
	},
}

var remoteAddCmd = &cobra.Command{
	Use:   "add name url",
	Args:  cobra.ExactArgs(2),
	Short: "Add a new data remote",
	Long:  "Add a new data remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		url := args[1]
		useLocal, _ := cmd.Flags().GetBool("local")
		cfg, err := utils.ReadConfig(useLocal)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
		}
		return addRemote(name, url, cfg)
	},
}

var remoteDefaultCmd = &cobra.Command{
	Use:   "default name",
	Args:  cobra.ExactArgs(1),
	Short: "Set/unset the default data remote",
	Long:  "Set/unset the default data remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		useLocal, _ := cmd.Flags().GetBool("local")
		cfg, err := utils.ReadConfig(useLocal)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
		}
		return setDefaultRemote(name, cfg)
	},
}

var remoteModifyCmd = &cobra.Command{
	Use:   "modify name option [value]",
	Args:  cobra.MinimumNArgs(2),
	Short: "Modify the configuration of a data remote",
	Long:  "Modify the configuration of a data remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		opt := args[1]
		val := ""
		if len(args) == 3 {
			val = args[2]
		}
		useLocal, _ := cmd.Flags().GetBool("local")
		cfg, err := utils.ReadConfig(useLocal)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
		}
		return modifyRemote(name, opt, val, cfg, useLocal)
	},
}

var remoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "Setup and manage data remotes",
	Long:  "Setup and manage data remotes",
	RunE: func(cmd *cobra.Command, args []string) error {
		useLocal, _ := cmd.Flags().GetBool("local")
		cfg, err := utils.ReadConfig(useLocal)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
		}
		return listRemotes(cfg)
	},
}

var remoteRemoveCmd = &cobra.Command{
	Use:   "remove name",
	Args:  cobra.ExactArgs(1),
	Short: "Remove a data remote",
	Long:  "Remove a data remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		useLocal, _ := cmd.Flags().GetBool("local")
		cfg, err := utils.ReadConfig(useLocal)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
		}
		return removeRemote(args[0], cfg)
	},
}

var remoteRenameCmd = &cobra.Command{
	Use:   "rename name new-name",
	Short: "Rename a data remote",
	Long:  "Rename a data remote",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("remote rename called")
	},
}

func addRemote(name, url string, cfg *utils.Config) error {
	if _, err := cfg.File.GetSection(fmt.Sprintf(`remote "%s"`, name)); err != nil {
		return errors.New(fmt.Sprintf(
			"Remote %s already exists, use remote modify to edit remote configuration",
			name,
		))
	}
	sec, _ := cfg.File.NewSection(fmt.Sprintf(`remote "%s"`, name))
	sec.NewKey("url", url)
	if err := cfg.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error saving %s: %s", cfg.Path, err))
	}
	return nil
}

func listRemotes(cfg *utils.Config) error {
	re := regexp.MustCompile(`remote\s+"([^"]*)"`)
	for _, sec := range cfg.File.Sections() {
		name := sec.Name()
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			url, _ := sec.GetKey("url")
			fmt.Println(matches[1], url)
		}
	}
	return nil
}
func modifyRemote(name, opt, val string, cfg *utils.Config, useLocal bool) error {
	sec, err := cfg.File.GetSection(fmt.Sprintf(`remote "%s"`, name))
	if sec == nil && !useLocal {
		return errors.New(fmt.Sprintf(
			"Remote %s not found in %s, use remote add %s to create\n",
			name,
			cfg.Path,
			name,
		))
	} else {
		sec, err = cfg.File.NewSection(fmt.Sprintf(`remote "%s"`, name))
		if err != nil {
			return errors.New(
				fmt.Sprintf("Error creating section in local config: %s", err),
			)
		}
	}
	key, err := sec.GetKey(opt)
	if err != nil {
		if val != "" {
			sec.NewKey(opt, val)
		}
		if err = cfg.Save(); err != nil {
			return errors.New(fmt.Sprintf("Error saving %s: %s\n", cfg.Path, err))
		}
		return nil
	}
	if val == "" {
		sec.DeleteKey(opt)
	} else {
		key.SetValue(val)
	}
	if err = cfg.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error saving %s: %s\n", cfg.Path, err))
	}
	return nil
}
func removeRemote(name string, cfg *utils.Config) error {
	cfg.File.DeleteSection(fmt.Sprintf(`remote "%s"`, name))
	if err := cfg.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error saving %s: %s\n", cfg.Path, err))
	}
	return nil
}

func setDefaultRemote(name string, cfg *utils.Config) error {
	sec, _ := cfg.File.GetSection("core")
	key, err := sec.GetKey("remote")
	if err != nil {
		sec.NewKey("remote", name)
	} else {
		key.SetValue(name)
	}
	if err = cfg.Save(); err != nil {
		return errors.New(fmt.Sprintf("Error saving %s: %s", cfg.Path, err))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(remoteCmd)

	remoteCmd.AddCommand(remoteAddCmd)
	remoteCmd.AddCommand(remoteDefaultCmd)
	remoteCmd.AddCommand(remoteModifyCmd)
	remoteCmd.AddCommand(remoteListCmd)
	remoteCmd.AddCommand(remoteRemoveCmd)
	remoteCmd.AddCommand(remoteRenameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")
	remoteCmd.PersistentFlags().Bool("local", false, "Use local config (.sdvc/config.local)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
