package manifest

import (
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	in := "url,status,must_delete\nhttps://example.com/a,pending,false\n"
	rows, err := Read(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].URL != "https://example.com/a" || rows[0].Status != "pending" || rows[0].MustDelete {
		t.Fatalf("row: %+v", rows[0])
	}
}

func TestReadMustDelete(t *testing.T) {
	in := "url,status,must_delete\nhttps://x,b,true\n"
	rows, err := Read(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if !rows[0].MustDelete {
		t.Fatal("expected MustDelete")
	}
}
