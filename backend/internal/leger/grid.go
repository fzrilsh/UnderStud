package leger

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// GridFromCSV membaca CSV jadi Grid. Baris kosong (semua sel kosong) dilewati.
func GridFromCSV(r io.Reader) (Grid, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	var rows [][]string
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Grid{}, fmt.Errorf("baca csv: %w", err)
		}
		if allEmpty(rec) {
			continue
		}
		rows = append(rows, rec)
	}
	return Grid{Rows: rows}, nil
}

// GridFromXLSX membaca sheet pertama .xlsx jadi Grid. Merged cell → nilai sel
// kiri-atas; sel kosong → "". Angka diformat tanpa ".0".
func GridFromXLSX(r io.Reader) (Grid, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return Grid{}, fmt.Errorf("buka xlsx: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return Grid{}, fmt.Errorf("xlsx tanpa sheet")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return Grid{}, fmt.Errorf("baca sheet: %w", err)
	}
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		norm := make([]string, len(row))
		for i, c := range row {
			norm[i] = strings.TrimSpace(c)
		}
		out = append(out, norm)
	}
	return Grid{Rows: out}, nil
}

func allEmpty(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// parseNilai: "85" atau "85.5" → 85.0; bukan angka → (0,false).
// F16: tolak nilai non-finite ("inf"/"NaN") — sel rusak tak boleh lolos ke
// payload (json.Marshal menolak Inf/NaN → payload NULL → preview 500).
func parseNilai(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}
