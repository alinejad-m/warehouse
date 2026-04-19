package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"warehouse/config"
	"warehouse/internal/gitrepo"
)

var dotenvPath string

func bindConfigFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&dotenvPath, "config", "", "Path to .env (optional; env vars may be set in the shell)")
}

func loadCfg() (*config.Config, error) {
	path := dotenvPath
	if path == "" {
		path = ".env"
	}
	return config.Load(path)
}

func openRepo(cfg *config.Config) (*gitrepo.Repo, error) {
	if err := config.ResolveGitWorkDir(cfg); err != nil {
		return nil, err
	}
	repo := &gitrepo.Repo{
		WorkDir: cfg.WorkDir,
		Branch:  cfg.GitBranch,
		Remote:  cfg.GitRemote,
		Author:  cfg.CommitAuthor,
		Email:   cfg.CommitEmail,
	}
	if !repo.Exists() {
		return nil, fmt.Errorf("no git repository at %q", cfg.WorkDir)
	}
	return repo, nil
}

func manifestAbs(cfg *config.Config) string {
	return filepath.Join(cfg.WorkDir, filepath.FromSlash(cfg.ManifestPath))
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.CreateTemp(filepath.Dir(dst), ".cp-*")
	if err != nil {
		return err
	}
	tmpPath := out.Name()
	n, err := io.Copy(out, in)
	if err != nil {
		out.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	_ = n
	if err := os.Rename(tmpPath, dst); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func ensureRepoReady(repo *gitrepo.Repo) error {
	if !repo.Exists() {
		return fmt.Errorf("repository missing at %q; run `warehouse init` first", repo.WorkDir)
	}
	return nil
}
