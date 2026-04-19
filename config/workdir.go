package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveGitWorkDir sets cfg.WorkDir to an absolute path that contains .git:
// - uses cfg.WorkDir if it is already a git root;
// - if WAREHOUSE_WORKDIR was not set in the environment, walks upward from the process cwd;
// - if WAREHOUSE_WORKDIR was set explicitly but is not a git repo, returns an error (no silent cwd steal).
func ResolveGitWorkDir(cfg *Config) error {
	if ok, err := IsGitWorkDir(cfg.WorkDir); err != nil {
		return err
	} else if ok {
		abs, err := filepath.Abs(cfg.WorkDir)
		if err != nil {
			return fmt.Errorf("workdir abs: %w", err)
		}
		cfg.WorkDir = abs
		return nil
	}
	if cfg.GitWorkDirExplicit {
		return fmt.Errorf("WAREHOUSE_WORKDIR %q has no .git — mount your repository here (with .git) or unset WAREHOUSE_WORKDIR so the tool can find a repo from the current directory", cfg.WorkDir)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if ok, _ := IsGitWorkDir(dir); ok {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("workdir abs: %w", err)
			}
			cfg.WorkDir = abs
			return nil
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return fmt.Errorf("no git repository: not at WAREHOUSE_WORKDIR %q and none found walking up from cwd %q", cfg.WorkDir, cwd)
}

// IsGitWorkDir reports whether dir contains a .git file or directory.
func IsGitWorkDir(dir string) (bool, error) {
	st, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return st.IsDir() || st.Mode().IsRegular(), nil
}

// WorkDirFromEnv reports whether WAREHOUSE_WORKDIR was set to a non-empty value in the environment.
func WorkDirFromEnv() bool {
	return strings.TrimSpace(os.Getenv("WAREHOUSE_WORKDIR")) != ""
}
