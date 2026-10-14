package leger

import "testing"

func sampleGrid() Grid {
	return Grid{Rows: [][]string{
		{"Sekolah", "", "SMKN 1 TANGERANG", ""},
		{"NPSN", "", "20606915", ""},
		{"Kelas", "", "X TJKT 1", ""},
		{"Tahun Pelajaran", "", "2025/2026 Genap", ""},
		{},
		{"No", "NAMA PESERTA DIDIK", "NISN", "Matematika (Umum)", "Bahasa Indonesia", "S", "I", "A"},
		{"1", "Andi Pratama", "2060691501", "85", "90", "3", "1", "0"},
		{"2", "Bunga Lestari", "2060691502", "78", "", "0", "0", "2"},
		{"3", "", "", "", "", "", "", ""}, // baris kosong → dilewati
	}}
}

func TestParseMetadata(t *testing.T) {
	l, err := Parse(sampleGrid())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if l.Sekolah != "SMKN 1 TANGERANG" || l.NPSN != "20606915" || l.Kelas != "X TJKT 1" || l.Periode != "2025/2026 Genap" {
		t.Fatalf("bad metadata: %+v", l)
	}
}

func TestParseRows(t *testing.T) {
	l, _ := Parse(sampleGrid())
	if len(l.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(l.Rows))
	}
	r := l.Rows[0]
	if r.NISN != "2060691501" || r.Nama != "Andi Pratama" {
		t.Fatalf("bad row: %+v", r)
	}
	if len(r.Nilai) != 2 || r.Nilai[0].Mapel != "Matematika (Umum)" || r.Nilai[0].Nilai != 85 {
		t.Fatalf("bad nilai: %+v", r.Nilai)
	}
	if r.Sakit != 3 || r.Izin != 1 || r.Alpa != 0 {
		t.Fatalf("bad absensi: %+v", r)
	}
}

func TestParseEmptyMapelSkipped(t *testing.T) {
	l, _ := Parse(sampleGrid())
	if len(l.Rows[1].Nilai) != 1 {
		t.Fatalf("mapel kosong harus dilewati, got %+v", l.Rows[1].Nilai)
	}
}

func TestParseHeaderShifted(t *testing.T) {
	// header di baris berbeda → tetap terdeteksi (auto-detect).
	g := Grid{Rows: [][]string{
		{"laporan"},
		{"", ""},
		{"No", "NAMA PESERTA DIDIK", "NISN", "Fisika"},
		{"1", "Citra", "2060691503", "80"},
	}}
	l, err := Parse(g)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(l.Rows) != 1 || l.Rows[0].Nilai[0].Mapel != "Fisika" {
		t.Fatalf("header geser gagal: %+v", l.Rows)
	}
}

func TestParseNoHeader(t *testing.T) {
	_, err := Parse(Grid{Rows: [][]string{{"a", "b"}, {"c", "d"}}})
	if err == nil {
		t.Fatal("want error saat header tak ditemukan")
	}
}
