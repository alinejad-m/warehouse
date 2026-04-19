package pathutil

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path"
	"strings"
)

// HashHex returns the lowercase hex sha256 of the trimmed URL (stable file prefix under files/).
func HashHex(sourceURL string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(sourceURL)))
	return hex.EncodeToString(h[:])
}

// FileRelPath returns a stable repo-relative path under files/ for a source URL.
// extOverride is optional (e.g. from Content-Disposition); if empty, the URL path extension or ".bin" is used.
func FileRelPath(sourceURL string, extOverride string) string {
	prefix := HashHex(sourceURL)
	ext := strings.TrimSpace(extOverride)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	if ext == "" {
		ext = guessExtFromURL(sourceURL)
	}
	if ext == "" {
		ext = ".bin"
	}
	return path.Join("files", prefix+ext)
}

func guessExtFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	base := path.Base(u.Path)
	if base == "." || base == "/" {
		return ""
	}
	if i := strings.LastIndex(base, "."); i > 0 && i < len(base)-1 {
		ext := base[i:]
		if len(ext) <= 8 {
			return ext
		}
	}
	return ""
}
