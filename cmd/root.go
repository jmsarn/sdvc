/*
Copyright © 2025 James Arnold <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sdvc",
	Short: "A simple data version control utility",
	Long: `SDVC is a simple data version control utility inspired by DVC with a focus on
performance and a smaller featureset. Main features are managing references 
to data artifacts kept in cloud storage. The actual versioning of objects
is left to the cloud storage provider (e.g., S3) so ensure that the target
storage location has versioning enabled.`,
}

func toggleDebug(cmd *cobra.Command, _ []string) {
	debug, _ := cmd.Flags().GetBool("debug")
	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sdvc.yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
