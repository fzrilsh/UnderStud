package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"

	"backend/internal/db"
	"backend/internal/utils"
)

const exportDisclaimer = "Dokumen ini ringkasan berbantuan AI, bukan diagnosis. Keputusan tetap pada pertimbangan profesional Guru BK/Wali Kelas."

type ExportFile struct {
	Data        []byte
	ContentType string
	Filename    string
}

func renderTXT(studentName, kelas, content string, createdAt time.Time) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "RINGKASAN SISWA\n")
	fmt.Fprintf(&b, "Nama  : %s\n", studentName)
	fmt.Fprintf(&b, "Kelas : %s\n", kelas)
	fmt.Fprintf(&b, "Dibuat: %s WIB\n", createdAt.In(jakartaLoc()).Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "------------------------------------------------------------\n\n")
	b.WriteString(content)
	fmt.Fprintf(&b, "\n\n------------------------------------------------------------\n")
	fmt.Fprintf(&b, "%s\n", exportDisclaimer)
	return []byte(b.String())
}

func renderPDF(studentName, kelas, content string, createdAt time.Time) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "Ringkasan Siswa", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 6, "Nama  : "+sanitizePDFText(studentName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Kelas : "+sanitizePDFText(kelas), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Dibuat: "+createdAt.In(jakartaLoc()).Format("2006-01-02 15:04")+" WIB", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "", 11)
	body := content
	if strings.TrimSpace(body) == "" {
		body = "(Ringkasan kosong.)"
	}
	pdf.MultiCell(0, 6, sanitizePDFText(body), "", "L", false)
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "I", 9)
	pdf.MultiCell(0, 5, sanitizePDFText(exportDisclaimer), "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// sanitizePDFText: fpdf core font = Latin-1; ganti karakter di luar itu agar tak rusak.
func sanitizePDFText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || (r >= 32 && r <= 255) {
			b.WriteRune(r)
		} else {
			b.WriteRune('?')
		}
	}
	return b.String()
}

// summaryToPlainText: ubah content (objek JSON) jadi teks untuk PDF/TXT.
// Bila content bukan JSON (data lama), kembalikan apa adanya.
func summaryToPlainText(content string) string {
	var c struct {
		Headline  string `json:"headline"`
		Narrative string `json:"narrative"`
		Insights  []struct {
			Title string `json:"title"`
			Items []struct {
				Text string `json:"text"`
			} `json:"items"`
		} `json:"insights"`
	}
	if err := json.Unmarshal([]byte(content), &c); err != nil {
		return content
	}
	var b strings.Builder
	if strings.TrimSpace(c.Headline) != "" {
		b.WriteString(c.Headline)
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(c.Narrative) != "" {
		b.WriteString(c.Narrative)
		b.WriteString("\n\n")
	}
	for _, ins := range c.Insights {
		if strings.TrimSpace(ins.Title) == "" || len(ins.Items) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s\n", ins.Title)
		for _, it := range ins.Items {
			fmt.Fprintf(&b, "  - %s\n", it.Text)
		}
		b.WriteString("\n")
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return content
	}
	return out
}

// Export: generate PDF/TXT dari ringkasan + catat audit. Scoped ke kelas Wali Kelas.
func (s *Summary) Export(ctx context.Context, actor Actor, studentID, summaryID uuid.UUID, format string) (ExportFile, error) {
	if err := s.ensureScope(ctx, actor, studentID); err != nil {
		return ExportFile{}, err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format != "pdf" && format != "txt" {
		return ExportFile{}, utils.NewErrInvalidInput("Format ekspor harus pdf atau txt.")
	}
	sum, err := s.queries.GetSummaryByID(ctx, summaryID)
	if err != nil {
		return ExportFile{}, utils.NewErrNotFound("Ringkasan tidak ditemukan.")
	}
	if sum.StudentID != studentID {
		return ExportFile{}, utils.NewErrNotFound("Ringkasan tidak ditemukan.")
	}
	st, err := s.queries.GetStudentByID(ctx, studentID)
	if err != nil {
		return ExportFile{}, utils.NewErrNotFound("Siswa tidak ditemukan.")
	}

	body := summaryToPlainText(sum.Content)
	base := fmt.Sprintf("ringkasan-%s-%s", slugify(st.Name), sum.CreatedAt.In(jakartaLoc()).Format("20060102"))
	var out ExportFile
	switch format {
	case "pdf":
		data, err := renderPDF(st.Name, st.Kelas, body, sum.CreatedAt)
		if err != nil {
			return ExportFile{}, utils.NewErrInternal("gagal membuat PDF")
		}
		out = ExportFile{Data: data, ContentType: "application/pdf", Filename: base + ".pdf"}
	case "txt":
		out = ExportFile{Data: renderTXT(st.Name, st.Kelas, body, sum.CreatedAt), ContentType: "text/plain; charset=utf-8", Filename: base + ".txt"}
	}

	if _, err := s.queries.CreateExportRecord(ctx, db.CreateExportRecordParams{
		SummaryID: summaryID, ExportedBy: actor.UserID, Format: format,
	}); err != nil {
		return ExportFile{}, utils.NewErrInternal("gagal mencatat ekspor")
	}
	return out, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
