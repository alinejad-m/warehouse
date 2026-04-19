package manifest

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	HeaderURL        = "url"
	HeaderStatus     = "status"
	HeaderMustDelete = "must_delete"
)

// Row is one line in urls.csv (three columns).
type Row struct {
	URL        string
	Status     string
	MustDelete bool
}

// ReadFile parses the manifest CSV.
func ReadFile(path string) ([]Row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

// Read parses CSV from r. Expects header url,status,must_delete.
func Read(r io.Reader) ([]Row, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	header := records[0]
	if len(header) < 3 {
		return nil, errors.New("csv: need header url,status,must_delete")
	}
	idxURL := -1
	idxStatus := -1
	idxDel := -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case HeaderURL:
			idxURL = i
		case HeaderStatus:
			idxStatus = i
		case HeaderMustDelete, "must ldete", "must_ldete":
			idxDel = i
		}
	}
	if idxURL < 0 || idxStatus < 0 || idxDel < 0 {
		return nil, fmt.Errorf("csv: missing columns (got %v)", header)
	}
	var rows []Row
	for _, rec := range records[1:] {
		if len(rec) <= idxURL {
			continue
		}
		u := strings.TrimSpace(rec[idxURL])
		if u == "" {
			continue
		}
		st := ""
		if len(rec) > idxStatus {
			st = strings.TrimSpace(rec[idxStatus])
		}
		del := false
		if len(rec) > idxDel {
			del = parseBool(rec[idxDel])
		}
		rows = append(rows, Row{URL: u, Status: st, MustDelete: del})
	}
	return rows, nil
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "true" || s == "1" || s == "yes" || s == "y"
}

// WriteFile writes rows with the standard header.
func WriteFile(path string, rows []Row) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Write(f, rows)
}

// Write serializes rows to w.
func Write(w io.Writer, rows []Row) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{HeaderURL, HeaderStatus, HeaderMustDelete}); err != nil {
		return err
	}
	for _, row := range rows {
		del := "false"
		if row.MustDelete {
			del = "true"
		}
		if err := cw.Write([]string{row.URL, row.Status, del}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// AppendRow appends a single row to an existing file, creating parent dirs and header if missing.
func AppendRow(path string, row Row) error {
	dir := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		dir = path[:i]
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	_, statErr := os.Stat(path)
	newFile := os.IsNotExist(statErr)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if newFile {
		if _, err := f.WriteString("url,status,must_delete\n"); err != nil {
			return err
		}
	}
	cw := csv.NewWriter(f)
	del := "false"
	if row.MustDelete {
		del = "true"
	}
	if err := cw.Write([]string{row.URL, row.Status, del}); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}
