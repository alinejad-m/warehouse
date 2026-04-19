// Package config loads warehouse CLI settings from environment (and optional .env).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all settings for git sync and file downloads.
type Config struct {
	// GitURL is only used by `warehouse init` when the work directory is not yet a clone (optional if you clone manually).
	GitURL        string
	GitBranch     string // default: main
	GitRemote     string // remote name for pull/push, default: origin
	WorkDir       string // git working copy root (default: current directory)
	ManifestPath  string // path inside repo, default: manifest/urls.csv
	DownloadDir   string // staging for HTTP downloads (default: ./downloads); use an absolute path in Docker
	CommitAuthor  string // GIT_AUTHOR_NAME override
	CommitEmail   string // GIT_AUTHOR_EMAIL override
	SkipTLSVerify bool   // WAREHOUSE_HTTP_INSECURE_SKIP_VERIFY (dev only)
}

// Load reads optional .env from path (if the file exists), then builds Config from environment variables.
func Load(dotenvPath string) (*Config, error) {
	if dotenvPath != "" {
		if _, err := os.Stat(dotenvPath); err == nil {
			_ = godotenv.Load(dotenvPath)
		}
	}
	dl := getenvDefault("WAREHOUSE_DOWNLOAD_DIR", "downloads")
	if !filepath.IsAbs(dl) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getwd: %w", err)
		}
		dl = filepath.Join(wd, dl)
	}
	dl = filepath.Clean(dl)

	wd := getenvDefault("WAREHOUSE_WORKDIR", ".")
	if !filepath.IsAbs(wd) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getwd: %w", err)
		}
		wd = filepath.Join(cwd, wd)
	}
	wd = filepath.Clean(wd)

	cfg := &Config{
		GitURL:       os.Getenv("WAREHOUSE_GIT_URL"),
		GitBranch:    getenvDefault("WAREHOUSE_GIT_BRANCH", "main"),
		GitRemote:    getenvDefault("WAREHOUSE_GIT_REMOTE", "origin"),
		WorkDir:      wd,
		ManifestPath: getenvDefault("WAREHOUSE_MANIFEST_PATH", "manifest/urls.csv"),
		DownloadDir:  dl,
		CommitAuthor: getenvDefault("WAREHOUSE_GIT_AUTHOR_NAME", "warehouse-bot"),
		CommitEmail:  getenvDefault("WAREHOUSE_GIT_AUTHOR_EMAIL", "warehouse-bot@local"),
		SkipTLSVerify: os.Getenv("WAREHOUSE_HTTP_INSECURE_SKIP_VERIFY") == "true" ||
			os.Getenv("WAREHOUSE_HTTP_INSECURE_SKIP_VERIFY") == "1",
	}
	return cfg, nil
}

func getenvDefault(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
