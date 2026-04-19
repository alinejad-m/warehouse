package manifest

import "strings"

const (
	StatusPending    = "pending"
	StatusDownloaded = "downloaded"
	StatusFailed     = "failed"
	StatusRemoved    = "removed"
)

// NormalizeStatus lowercases and trims a status cell.
func NormalizeStatus(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
