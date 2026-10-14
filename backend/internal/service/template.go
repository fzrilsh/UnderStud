package service

import (
	"bytes"
	"encoding/csv"
	"strings"

	"github.com/xuri/excelize/v2"

	"backend/internal/utils"
)

// contoh kolom mapel + absensi untuk template.
var templateMapel = []string{"Matematika (Umum)", "Bahasa Indonesia", "Informatika"}

// BuildTemplate membuat file leger kosong (format xlsx|csv) dengan kelas WK
// di-prefill. Struktur identik dengan yang dibaca leger.Parse (spec §10.5).
func BuildTemplate(format, kelas string) ([]byte, string, string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "xlsx"
	}
	meta := [][2]string{
		{"Sekolah", ""},
		{"NPSN", ""},
		{"Kelas", kelas},
		{"Tahun Pelajaran", ""},
	}
	header := append([]string{"No", "NAMA PESERTA DIDIK", "NISN"}, templateMapel...)
	header = append(header, "S", "I", "A")
	contoh := []string{"1", "CONTOH — hapus sebelum upload", "0000000000", "85", "80", "88", "0", "0", "0"}

	switch format {
	case "csv":
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		for _, m := range meta {
			_ = w.Write([]string{m[0], "", m[1]})
		}
		_ = w.Write([]string{})
		_ = w.Write(header)
		_ = w.Write(contoh)
		w.Flush()
		return buf.Bytes(), "text/csv; charset=utf-8", "template-leger-" + slugify(kelas) + ".csv", nil
	case "xlsx":
		f := excelize.NewFile()
		sheet := f.GetSheetName(0)
		row := 1
		for _, m := range meta {
			_ = f.SetCellValue(sheet, cell("A", row), m[0])
			_ = f.SetCellValue(sheet, cell("C", row), m[1])
			row++
		}
		row++ // baris kosong
		for i, h := range header {
			_ = f.SetCellValue(sheet, colName(i)+itoa(row), h)
		}
		row++
		for i, v := range contoh {
			_ = f.SetCellValue(sheet, colName(i)+itoa(row), v)
		}
		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return nil, "", "", utils.NewErrInternal("gagal membuat template")
		}
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "template-leger-" + slugify(kelas) + ".xlsx", nil
	}
	return nil, "", "", utils.NewErrInvalidInput("Format template harus xlsx atau csv.")
}

func colName(i int) string {
	name := ""
	for i >= 0 {
		name = string(rune('A'+i%26)) + name
		i = i/26 - 1
	}
	return name
}

func cell(col string, row int) string { return col + itoa(row) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
