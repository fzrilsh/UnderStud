package service

import "testing"

func TestParseCSVGroupsByNIS(t *testing.T) {
	csvData := "nis,nama,kelas,mapel,nilai,periode\n" +
		"2026001,Andi,XI-IPA-1,Matematika,85,2026-Ganjil\n" +
		"2026001,Andi,XI-IPA-1,Fisika,78,2026-Ganjil\n" +
		"2026002,Bunga,XI-IPA-1,Matematika,90,2026-Ganjil\n"
	rows, errs := parseCSV([]byte(csvData), "XI-IPA-1")
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 students, got %d", len(rows))
	}
	if len(rows[0].Nilai) != 2 {
		t.Fatalf("want 2 nilai for 2026001, got %d", len(rows[0].Nilai))
	}
}

func TestParseCSVRejectsBadKelas(t *testing.T) {
	csvData := "nis,nama,kelas,mapel,nilai,periode\n" +
		"2026001,Andi,XI-IPA-2,Matematika,85,2026-Ganjil\n"
	_, errs := parseCSV([]byte(csvData), "XI-IPA-1")
	if len(errs) == 0 {
		t.Fatal("want error for out-of-scope kelas")
	}
}

func TestParseCSVRejectsBadNIS(t *testing.T) {
	csvData := "nis,nama,kelas,mapel,nilai,periode\n" +
		"AB@#$%,Andi,XI-IPA-1,Matematika,85,2026-Ganjil\n"
	_, errs := parseCSV([]byte(csvData), "XI-IPA-1")
	if len(errs) == 0 {
		t.Fatal("want error for invalid NIS")
	}
}

func TestParseCSVCaseInsensitiveHeader(t *testing.T) {
	csvData := "NIS,Nama,Kelas,Mapel,Nilai,Periode\n" +
		"2026001,Andi,XI-IPA-1,Matematika,85,2026-Ganjil\n"
	rows, errs := parseCSV([]byte(csvData), "XI-IPA-1")
	if len(errs) != 0 || len(rows) != 1 {
		t.Fatalf("header should be case-insensitive: rows=%d errs=%v", len(rows), errs)
	}
}
