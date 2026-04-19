package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"warehouse/internal/manifest"
)

var addMustDelete bool

var addCmd = &cobra.Command{
	Use:   "add [url]",
	Short: "Append a URL row to the CSV (status=pending), commit, and push",
	Args:  cobra.ExactArgs(1),
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
		if err := os.MkdirAll(cfg.DownloadDir, 0o755); err != nil {
			return err
		}
		mPath := manifestAbs(cfg)
		if err := os.MkdirAll(filepath.Dir(mPath), 0o755); err != nil {
			return err
		}
		row := manifest.Row{
			URL:        args[0],
			Status:     manifest.StatusPending,
			MustDelete: addMustDelete,
		}
		if err := manifest.AppendRow(mPath, row); err != nil {
			return err
		}
		relMan := strings.ReplaceAll(cfg.ManifestPath, `\`, "/")
		if err := repo.Add(relMan); err != nil {
			return err
		}
		if err := repo.Commit(fmt.Sprintf("manifest: add %s", args[0])); err != nil {
			return err
		}
		if err := repo.Push(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "add: pushed")
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&addMustDelete, "must-delete", false, "Set must_delete=true for this URL (cleanup command removes file + row later)")
}
