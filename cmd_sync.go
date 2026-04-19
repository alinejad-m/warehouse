package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run git pull in the working copy (fetch latest manifest and files from remote)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadCfg()
		if err != nil {
			return err
		}
		repo, err := openRepo(cfg)
		if err != nil {
			return err
		}
		if err := ensureRepoReady(repo); err != nil {
			return err
		}
		if err := repo.Pull(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "pull: ok")
		return nil
	},
}
