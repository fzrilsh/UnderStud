// Package leger mem-parsing file Leger Nilai Akhir dari E-Rapor
// (matriks lebar: 1 baris = 1 siswa, tiap mapel 1 kolom). Spec §10.
package leger

// Grid = matriks string 2D netral; sumber bisa .xlsx atau .csv.
type Grid struct {
	Rows [][]string
}

// NilaiRow = satu nilai mapel.
type NilaiRow struct {
	Mapel string
	Nilai float64
}

// LegerRow = satu siswa dalam leger.
type LegerRow struct {
	NISN  string
	Nama  string
	Nilai []NilaiRow
	Sakit int
	Izin  int
	Alpa  int
}

// Leger = hasil parse lengkap.
type Leger struct {
	Sekolah string
	NPSN    string
	Kelas   string
	Periode string
	Rows    []LegerRow
}
