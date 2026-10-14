package leger

import (
	"fmt"
	"strings"
)

// Parse mengubah Grid jadi Leger. Metadata & baris header dideteksi otomatis
// (tahan kalau E-Rapor beda versi menggeser baris). Spec §10.1/§10.4.
func Parse(g Grid) (Leger, error) {
	l := Leger{}

	// 1. Metadata: cari sel berlabel, ambil nilai di kolom terdekat yang terisi.
	l.Sekolah = findMeta(g, "sekolah")
	l.NPSN = findMeta(g, "npsn")
	l.Kelas = findMeta(g, "kelas")
	l.Periode = findMeta(g, "tahun pelajaran")

	// 2. Header: baris pertama yang memuat NAMA dan NISN.
	headerRow, namaCol, nisnCol := -1, -1, -1
	for i, row := range g.Rows {
		nc, ic := -1, -1
		for j, cell := range row {
			u := strings.ToUpper(strings.TrimSpace(cell))
			if strings.Contains(u, "NAMA") {
				nc = j
			}
			if u == "NISN" {
				ic = j
			}
		}
		if nc >= 0 && ic >= 0 {
			headerRow, namaCol, nisnCol = i, nc, ic
			break
		}
	}
	if headerRow < 0 {
		return Leger{}, fmt.Errorf("format leger tidak dikenali (baris header NAMA/NISN tidak ditemukan)")
	}

	header := g.Rows[headerRow]
	// Kolom mapel = semua kolom selain No/NAMA/NISN/S/I/A.
	absensiCol := map[string]int{}
	var mapelCols []int
	for j, cell := range header {
		u := strings.ToUpper(strings.TrimSpace(cell))
		switch {
		case j == namaCol || j == nisnCol:
			continue
		case u == "NO" || u == "":
			continue
		case u == "S" || u == "I" || u == "A":
			absensiCol[u] = j
		default:
			mapelCols = append(mapelCols, j)
		}
	}

	// 3. Baris data.
	for i := headerRow + 1; i < len(g.Rows); i++ {
		row := g.Rows[i]
		nisn := cellAt(row, nisnCol)
		nama := cellAt(row, namaCol)
		if nisn == "" && nama == "" {
			continue // baris kosong (template belum diisi)
		}
		lr := LegerRow{NISN: nisn, Nama: nama}
		for _, j := range mapelCols {
			raw := cellAt(row, j)
			if raw == "" {
				continue
			}
			v, ok := parseNilai(raw)
			if !ok {
				continue
			}
			lr.Nilai = append(lr.Nilai, NilaiRow{Mapel: strings.TrimSpace(header[j]), Nilai: v})
		}
		if j, ok := absensiCol["S"]; ok {
			lr.Sakit = atoiSafe(cellAt(row, j))
		}
		if j, ok := absensiCol["I"]; ok {
			lr.Izin = atoiSafe(cellAt(row, j))
		}
		if j, ok := absensiCol["A"]; ok {
			lr.Alpa = atoiSafe(cellAt(row, j))
		}
		l.Rows = append(l.Rows, lr)
	}
	return l, nil
}

// findMeta: cari sel yang memuat label (case-insensitive), ambil sel terisi
// berikutnya di baris yang sama.
func findMeta(g Grid, label string) string {
	for _, row := range g.Rows {
		for j, cell := range row {
			if strings.Contains(strings.ToLower(strings.TrimSpace(cell)), label) {
				for k := j + 1; k < len(row); k++ {
					if v := strings.TrimSpace(row[k]); v != "" {
						return v
					}
				}
			}
		}
	}
	return ""
}

func cellAt(row []string, j int) string {
	if j < 0 || j >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[j])
}

func atoiSafe(s string) int {
	n := 0
	for _, r := range strings.TrimSpace(s) {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
