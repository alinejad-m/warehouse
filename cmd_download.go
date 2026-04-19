package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"warehouse/internal/download"
	"warehouse/internal/manifest"
	"warehouse/internal/pathutil"
)

var downloadNoPull bool

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Pull, download every pending URL into the mount dir, copy into the repo, update status, commit, and push",
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
		if !downloadNoPull {
			if err := repo.Pull(); err != nil {
				return err
			}
		}
		if err := os.MkdirAll(cfg.DownloadDir, 0o755); err != nil {
			return err
		}
		mPath := manifestAbs(cfg)
		rows, err := manifest.ReadFile(mPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		opt := download.Options{InsecureSkipVerify: cfg.SkipTLSVerify}
		changed := false
		for i := range rows {
			if manifest.NormalizeStatus(rows[i].Status) != manifest.StatusPending {
				continue
			}
			u := rows[i].URL
			stage := filepath.Join(cfg.DownloadDir, pathutil.HashHex(u)+".part")
			_ = os.Remove(stage)
			_, extHint, derr := download.ToFile(u, stage, opt)
			if derr != nil {
				rows[i].Status = manifest.StatusFailed
				changed = true
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "download fail %s: %v\n", u, derr)
				continue
			}
			rel := pathutil.FileRelPath(u, extHint)
			dest := filepath.Join(repo.WorkDir, filepath.FromSlash(rel))
			matches, _ := pathutil.GlobRepoFilesForURL(repo.WorkDir, u)
			for _, m := range matches {
				if filepath.Clean(m) == filepath.Clean(dest) {
					continue
				}
				relOld, er := filepath.Rel(repo.WorkDir, m)
				if er == nil {
					_ = repo.Remove(true, relOld)
				}
				_ = os.Remove(m)
			}
			if err := copyFile(stage, dest); err != nil {
				rows[i].Status = manifest.StatusFailed
				changed = true
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "copy fail %s: %v\n", u, err)
				_ = os.Remove(stage)
				continue
			}
			_ = os.Remove(stage)
			rows[i].Status = manifest.StatusDownloaded
			changed = true
			relGit := strings.ReplaceAll(rel, `\`, "/")
			if err := repo.Add(relGit); err != nil {
				return err
			}
		}
		if !changed {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "download: nothing pending")
			return nil
		}
		if err := manifest.WriteFile(mPath, rows); err != nil {
			return err
		}
		if err := repo.Add(strings.ReplaceAll(cfg.ManifestPath, `\`, "/")); err != nil {
			return err
		}
		if err := repo.Commit("download: update manifest and files"); err != nil {
			return err
		}
		if err := repo.Push(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "download: pushed")
		return nil
	},
}

func init() {
	downloadCmd.Flags().BoolVar(&downloadNoPull, "no-pull", false, "Do not run git pull before downloading")
}
