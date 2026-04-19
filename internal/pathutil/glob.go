package pathutil

import (
	"path/filepath"
)

// GlobRepoFilesForURL returns on-disk paths under workDir for files matching this URL's content hash (files/<hex>*).
func GlobRepoFilesForURL(workDir, sourceURL string) ([]string, error) {
	pat := filepath.Join(workDir, "files", HashHex(sourceURL)+"*")
	return filepath.Glob(pat)
}
