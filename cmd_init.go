package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"warehouse/internal/gitrepo"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Prepare dirs; if WAREHOUSE_WORKDIR is already a git clone, use its configured remote (no URL env needed). Otherwise clone using WAREHOUSE_GIT_URL once",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadCfg()
		if err != nil {
			return err
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		repo := &gitrepo.Repo{
			WorkDir: cfg.WorkDir,
			Branch:  cfg.GitBranch,
			Remote:  cfg.GitRemote,
			Author:  cfg.CommitAuthor,
			Email:   cfg.CommitEmail,
		}
		if repo.Exists() {
			if u, e := gitrepo.GetRemoteURL(cfg.WorkDir, cfg.GitRemote); e == nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "git remote %q: %s\n", cfg.GitRemote, u)
			}
			if err := os.MkdirAll(cfg.DownloadDir, 0o755); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ready: %s (branch %s)\n", repo.WorkDir, repo.Branch)
			return nil
		}
		if cfg.GitURL == "" {
			return fmt.Errorf("no .git at %q — clone this repo here (git clone …) or set WAREHOUSE_GIT_URL and WAREHOUSE_WORKDIR to a new empty folder for automated clone", cfg.WorkDir)
		}
		if filepath.Clean(cfg.WorkDir) == filepath.Clean(cwd) {
			return fmt.Errorf("refusing to clone into the current working directory %q; set WAREHOUSE_WORKDIR to a new path (e.g. ./repo) or clone manually with git", cwd)
		}
		repo, err = gitrepo.InitOrClone(cfg.GitURL, cfg.WorkDir, cfg.GitBranch, cfg.GitRemote, cfg.CommitAuthor, cfg.CommitEmail)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(cfg.DownloadDir, 0o755); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cloned: %s (branch %s)\n", repo.WorkDir, repo.Branch)
		return nil
	},
}
