package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"warehouse/internal/gitrepo"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Show the git remote used for pull/push (same as git remote -v in WAREHOUSE_WORKDIR)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadCfg()
		if err != nil {
			return err
		}
		if _, err := openRepo(cfg); err != nil { // ensures WAREHOUSE_WORKDIR is a git clone
			return err
		}
		c := exec.Command("git", "remote", "-v")
		c.Dir = cfg.WorkDir
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git remote -v: %w\n%s", err, strings.TrimSpace(string(out)))
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), string(out))
		url, err := gitrepo.GetRemoteURL(cfg.WorkDir, cfg.GitRemote)
		if err == nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n# pull/push use remote %q → %s\n", cfg.GitRemote, url)
		}
		return nil
	},
}
