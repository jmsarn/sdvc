/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmsarn/sdvc/utils"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config name [value]",
	Short: "Get or set config options",
	Long: `Get or set the given configuration value:

sdvc config core.remote new-default-remote`,
	PreRun: toggleVerbose,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value := ""
		if len(args) == 2 {
			value = args[1]
		}
		useLocal, _ := cmd.Flags().GetBool("local")
		unset, _ := cmd.Flags().GetBool("unset")
		cfg, err := utils.ReadConfig(useLocal)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading %s: %s", cfg.Path, err))
		}
		return updateConfig(name, value, cfg, unset)
	},
}

func updateConfig(name, value string, cfg *utils.Config, unset bool) error {
	split := strings.Split(name, ".")
	sec, _ := cfg.File.GetSection(split[0])
	key, err := sec.GetKey(split[1])
	if value == "" {
		if err != nil {
			return errors.New(
				fmt.Sprintf("Key %s not found in config section %s\n", split[1], split[0]),
			)
		}
	} else {
		if unset {
			if err == nil {
				sec.DeleteKey(split[1])
			}
		} else {
			if err != nil {
				sec.NewKey(split[1], value)
			} else {
				key.SetValue(value)
			}
		}
		if err = cfg.Save(); err != nil {
			return errors.New(fmt.Sprintf("Error saving %s: %s\n", cfg.Path, err))
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	configCmd.Flags().Bool("local", false, "Use local config (.dvc/config.local)")
	configCmd.Flags().BoolP("unset", "u", false, "Unset option")
}
