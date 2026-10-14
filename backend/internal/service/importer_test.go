package service

import "testing"

func TestDetectFormat(t *testing.T) {
	cases := map[string]string{
		"leger.xlsx": "xlsx",
		"leger.XLSX": "xlsx",
		"leger.csv":  "csv",
		"leger.txt":  "",
		"data":       "",
	}
	for name, want := range cases {
		if got := detectFormat(name, nil); got != want {
			t.Errorf("detectFormat(%q)=%q want %q", name, got, want)
		}
	}
	if detectFormat("data", []byte{'P', 'K', 0x03, 0x04}) != "xlsx" {
		t.Error("magic bytes zip harus terdeteksi xlsx")
	}
}
