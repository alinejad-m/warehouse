package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"warehouse/internal/manifest"
	"warehouse/internal/pathutil"
)

var cleanupNoPull bool

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "For rows with must_delete=true, remove matching files from git, drop the row, commit, push, and delete staging files under the download dir",
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
		if !cleanupNoPull {
			if err := repo.Pull(); err != nil {
				return err
			}
		}
		mPath := manifestAbs(cfg)
		rows, err := manifest.ReadFile(mPath)
		if err != nil {
			if os.IsNotExist(err) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "cleanup: no manifest")
				return nil
			}
			return err
		}
		var kept []manifest.Row
		changed := false
		for _, row := range rows {
			if !row.MustDelete {
				kept = append(kept, row)
				continue
			}
			st := manifest.NormalizeStatus(row.Status)
			if st == manifest.StatusPending {
				changed = true
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cleanup: drop pending row %s\n", row.URL)
				continue
			}
			matches, _ := pathutil.GlobRepoFilesForURL(repo.WorkDir, row.URL)
			for _, abs := range matches {
				rel, er := filepath.Rel(repo.WorkDir, abs)
				if er != nil {
					continue
				}
				rel = strings.ReplaceAll(rel, `\`, "/")
				if err := repo.Remove(true, rel); err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "git rm %s: %v\n", rel, err)
				}
				_ = os.Remove(abs)
				changed = true
			}
			stage := filepath.Join(cfg.DownloadDir, pathutil.HashHex(row.URL)+".part")
			_ = os.Remove(stage)
			for _, g := range mustGlob(filepath.Join(cfg.DownloadDir, pathutil.HashHex(row.URL)+"*")) {
				_ = os.Remove(g)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cleanup: removed %s\n", row.URL)
			changed = true
			continue
		}
		if !changed {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "cleanup: nothing to do")
			return nil
		}
		if err := manifest.WriteFile(mPath, kept); err != nil {
			return err
		}
		if err := repo.Add(strings.ReplaceAll(cfg.ManifestPath, `\`, "/")); err != nil {
			return err
		}
		if err := repo.Commit("cleanup: remove must_delete rows and files"); err != nil {
			return err
		}
		if err := repo.Push(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "cleanup: pushed")
		return nil
	},
}

func mustGlob(pat string) []string {
	out, _ := filepath.Glob(pat)
	return out
}

func init() {
	cleanupCmd.Flags().BoolVar(&cleanupNoPull, "no-pull", false, "Do not run git pull before cleanup")
}
