package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var listPull bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print the manifest CSV (url, status, must_delete)",
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
		if listPull {
			if err := repo.Pull(); err != nil {
				return err
			}
		}
		mPath := manifestAbs(cfg)
		f, err := os.Open(mPath)
		if err != nil {
			if os.IsNotExist(err) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "(no manifest yet)")
				return nil
			}
			return err
		}
		defer f.Close()
		cr := csv.NewReader(f)
		recs, err := cr.ReadAll()
		if err != nil {
			return err
		}
		w := csv.NewWriter(cmd.OutOrStdout())
		for _, row := range recs {
			if err := w.Write(row); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()
	},
}

func init() {
	listCmd.Flags().BoolVar(&listPull, "pull", false, "Run git pull before reading the CSV")
}
