// Package config loads warehouse CLI settings from environment (and optional .env).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all settings for git sync and file downloads.
type Config struct {
	GitURL        string // full clone URL (may embed token), required for init
	GitBranch     string // default: main
	WorkDir       string // git working copy root
	ManifestPath  string // path inside repo, default: manifest/urls.csv
	DownloadDir   string // mounted volume for temporary download staging
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
	cfg := &Config{
		GitURL:       os.Getenv("WAREHOUSE_GIT_URL"),
		GitBranch:    getenvDefault("WAREHOUSE_GIT_BRANCH", "main"),
		WorkDir:      getenvDefault("WAREHOUSE_WORKDIR", filepath.Clean("warehouse-data/repo")),
		ManifestPath: getenvDefault("WAREHOUSE_MANIFEST_PATH", "manifest/urls.csv"),
		DownloadDir:  getenvDefault("WAREHOUSE_DOWNLOAD_DIR", filepath.Clean("warehouse-data/downloads")),
		CommitAuthor: getenvDefault("WAREHOUSE_GIT_AUTHOR_NAME", "warehouse-bot"),
		CommitEmail:  getenvDefault("WAREHOUSE_GIT_AUTHOR_EMAIL", "warehouse-bot@local"),
		SkipTLSVerify: os.Getenv("WAREHOUSE_HTTP_INSECURE_SKIP_VERIFY") == "true" ||
			os.Getenv("WAREHOUSE_HTTP_INSECURE_SKIP_VERIFY") == "1",
	}
	return cfg, nil
}

// ValidateGitURL returns an error if GitURL is empty.
func (c *Config) ValidateGitURL() error {
	if c.GitURL == "" {
		return fmt.Errorf("WAREHOUSE_GIT_URL is required")
	}
	return nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
