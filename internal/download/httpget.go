package download

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Options configures the HTTP client.
type Options struct {
	InsecureSkipVerify bool
}

// ToFile downloads remoteURL into destPath (parent dirs created). Returns bytes written and suggested extension (e.g. ".pdf") from Content-Disposition or URL.
func ToFile(remoteURL, destPath string, opt Options) (int64, string, error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return 0, "", err
	}
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: opt.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, remoteURL, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, "", fmt.Errorf("http %d", resp.StatusCode)
	}
	ext := extFromDisposition(resp.Header.Get("Content-Disposition"))
	if ext == "" {
		ext = guessExtFromURL(remoteURL)
	}
	tmp := destPath + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return 0, ext, err
	}
	n, err := io.Copy(f, resp.Body)
	cerr := f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return 0, ext, err
	}
	if cerr != nil {
		_ = os.Remove(tmp)
		return 0, ext, cerr
	}
	if err := os.Rename(tmp, destPath); err != nil {
		_ = os.Remove(tmp)
		return 0, ext, err
	}
	return n, ext, nil
}

func extFromDisposition(h string) string {
	// attachment; filename="book.pdf"
	h = strings.ToLower(h)
	if !strings.Contains(h, "filename") {
		return ""
	}
	parts := strings.Split(h, "filename")
	if len(parts) < 2 {
		return ""
	}
	v := strings.TrimSpace(parts[1])
	v = strings.TrimPrefix(v, "=")
	v = strings.TrimSpace(v)
	v = strings.Trim(v, `*"`)
	if i := strings.LastIndex(v, "."); i > 0 && i < len(v)-1 {
		return path.Ext(v)
	}
	return ""
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
	ext := path.Ext(base)
	if len(ext) > 8 {
		return ""
	}
	return ext
}
