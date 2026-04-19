package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "warehouse",
	Short:         "Sync a URL manifest CSV with git: add, download, push, and delete-by-flag",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the Cobra root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func init() {
	bindConfigFlag(rootCmd)
	rootCmd.AddCommand(initCmd, remoteCmd, syncCmd, addCmd, listCmd, downloadCmd, cleanupCmd)
}
