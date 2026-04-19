package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"warehouse/internal/gitrepo"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Clone WAREHOUSE_GIT_URL into WAREHOUSE_WORKDIR (idempotent if repo already exists)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadCfg()
		if err != nil {
			return err
		}
		if err := cfg.ValidateGitURL(); err != nil {
			return err
		}
		repo, err := gitrepo.InitOrClone(cfg.GitURL, cfg.WorkDir, cfg.GitBranch, cfg.CommitAuthor, cfg.CommitEmail)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(cfg.DownloadDir, 0o755); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ready: %s (branch %s)\n", repo.WorkDir, repo.Branch)
		return nil
	},
}
