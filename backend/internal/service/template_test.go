package service

import (
	"strings"
	"testing"
)

func TestBuildTemplateCSV(t *testing.T) {
	data, ct, fn, err := BuildTemplate("csv", "X TJKT 1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	s := string(data)
	for _, want := range []string{"Kelas", "X TJKT 1", "NAMA PESERTA DIDIK", "NISN"} {
		if !strings.Contains(s, want) {
			t.Fatalf("template csv missing %q:\n%s", want, s)
		}
	}
	if !strings.Contains(ct, "csv") || !strings.HasSuffix(fn, ".csv") {
		t.Fatalf("bad ct/fn: %s %s", ct, fn)
	}
}

func TestBuildTemplateXLSX(t *testing.T) {
	data, ct, fn, err := BuildTemplate("xlsx", "X TJKT 1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(data) < 2 || string(data[:2]) != "PK" {
		t.Fatalf("xlsx harus zip (PK)")
	}
	if !strings.Contains(ct, "spreadsheetml") || !strings.HasSuffix(fn, ".xlsx") {
		t.Fatalf("bad ct/fn: %s %s", ct, fn)
	}
}

func TestBuildTemplateBadFormat(t *testing.T) {
	if _, _, _, err := BuildTemplate("docx", "X"); err == nil {
		t.Fatal("want error untuk format tak didukung")
	}
}
