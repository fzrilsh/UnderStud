package ai

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// EraporRow: satu baris nilai e-rapor (untuk membangun struktur konten).
type EraporRow struct {
	Mapel   string  `json:"mapel"`
	Nilai   float64 `json:"nilai"`
	Periode string  `json:"periode"`
}

// SummaryContent = bentuk konten ringkasan yang dipakai frontend.
// Dibangun DETERMINISTIK dari data siswa; AI hanya mengisi Headline+Narrative.
type SummaryContent struct {
	Headline  string        `json:"headline"`
	Narrative string        `json:"narrative"`
	Metrics   []Metric      `json:"metrics"`
	Insights  []Insight     `json:"insights"`
	DataUsed  []DataUsedRow `json:"data_used"`
	RedFlag   *RedFlagNote  `json:"red_flag"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Insight struct {
	Icon     string        `json:"icon"`
	Title    string        `json:"title"`
	Subtitle string        `json:"subtitle"`
	Items    []InsightItem `json:"items"`
	Sources  []string      `json:"sources"`
	Muted    bool          `json:"muted"`
}

type InsightItem struct {
	Text string `json:"text"`
}

type DataUsedRow struct {
	Source  string `json:"source"`
	Details string `json:"details"`
	Updated string `json:"updated"`
	Status  string `json:"status"` // used | missing | not_sent
}

type RedFlagNote struct {
	Title string `json:"title"`
	Note  string `json:"note"`
}

// ContentInput: data terstruktur untuk membangun konten (di luar teks AI).
type ContentInput struct {
	StudentName          string
	Nilai                []EraporRow
	ObsCount             int
	RedFlagCount         int
	Categories           []string // label kategori indikator
	LastObsDate          time.Time
	HasLastObsDate       bool
	ProfileUpdated       time.Time
	MinatBakat           string
	PendapatWaliKelas    string
	MasukanBK            string
	CatatanKasusPrestasi string
	CatatanPsikologis    string
	AuthorName           string
	RedFlagTitle         string
}

var bulanSingkat = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

func fmtShortDate(t time.Time) string {
	return fmt.Sprintf("%d %s", t.Day(), bulanSingkat[int(t.Month())-1])
}

func nonEmpty(s string) bool { return strings.TrimSpace(s) != "" }

// BuildContent: susun struktur konten lengkap dari data siswa + headline/narrative AI.
func BuildContent(in ContentInput, headline, narrative string) SummaryContent {
	scores := append([]EraporRow{}, in.Nilai...)
	sort.SliceStable(scores, func(i, j int) bool { return scores[i].Nilai > scores[j].Nilai })

	avg := 0
	if len(scores) > 0 {
		sum := 0.0
		for _, s := range scores {
			sum += s.Nilai
		}
		avg = int(sum/float64(len(scores)) + 0.5)
	}
	var top, low *EraporRow
	if len(scores) > 0 {
		top = &scores[0]
		low = &scores[len(scores)-1]
	}
	hasRed := in.RedFlagCount > 0
	authorTitle := "Oleh " + in.AuthorName
	if !nonEmpty(in.AuthorName) {
		authorTitle = "Oleh Wali Kelas"
	}
	eraporDate := "Hari ini"
	if !in.ProfileUpdated.IsZero() {
		eraporDate = fmtShortDate(in.ProfileUpdated)
	}
	lastObsDate := "—"
	if in.HasLastObsDate {
		lastObsDate = fmtShortDate(in.LastObsDate)
	}

	// --- Insights: 4 kelompok (mirror desain) ---
	abilities := []InsightItem{}
	if top != nil {
		abilities = append(abilities, InsightItem{fmt.Sprintf("Nilai tertinggi di %s (%s).", top.Mapel, trimNum(top.Nilai))})
	}
	if len(scores) > 1 {
		abilities = append(abilities, InsightItem{fmt.Sprintf("Nilai %s (%s) juga menonjol.", scores[1].Mapel, trimNum(scores[1].Nilai))})
	}
	if nonEmpty(in.MinatBakat) {
		abilities = append(abilities, InsightItem{"Suka " + in.MinatBakat + "."})
	}
	if len(abilities) == 0 {
		abilities = append(abilities, InsightItem{"Belum ada kemampuan menonjol yang tercatat."})
	}

	strengths := []InsightItem{}
	if len(scores) > 0 {
		strengths = append(strengths, InsightItem{fmt.Sprintf("Rata-rata nilai %d dari %d mapel yang tercatat.", avg, len(scores))})
	}
	if in.ObsCount > 0 {
		strengths = append(strengths, InsightItem{fmt.Sprintf("Mengikuti %d sesi observasi kelas.", in.ObsCount)})
	}
	if len(strengths) == 0 {
		strengths = append(strengths, InsightItem{"Belum ada catatan kekuatan."})
	}

	needs := []InsightItem{}
	if len(in.Categories) > 0 {
		needs = append(needs, InsightItem{"Perlu dukungan di: " + strings.Join(in.Categories, ", ") + "."})
	}
	if low != nil && top != nil && low.Nilai < top.Nilai {
		needs = append(needs, InsightItem{fmt.Sprintf("Nilai %s (%s) paling rendah dan perlu perhatian.", low.Mapel, trimNum(low.Nilai))})
	}
	if hasRed {
		needs = append(needs, InsightItem{"Ada red flag terbuka — tindak lanjut bersama BK."})
	}
	if len(needs) == 0 {
		needs = append(needs, InsightItem{"Belum ada kebutuhan khusus yang ditandai."})
	}

	notEnough := []InsightItem{}
	if in.ObsCount < 2 {
		notEnough = append(notEnough, InsightItem{"Observasi masih sedikit — tren belum konklusif."})
	}
	if !nonEmpty(in.CatatanKasusPrestasi) {
		notEnough = append(notEnough, InsightItem{"Belum ada catatan kasus atau prestasi, jadi bagian ini tidak dibahas."})
	}
	notEnough = append(notEnough, InsightItem{"Catatan psikologis bersifat terbatas dan tidak pernah dikirim ke AI."})

	mkInsight := func(icon, title string, items []InsightItem, sources []string, muted bool) Insight {
		return Insight{
			Icon: icon, Title: title,
			Subtitle: fmt.Sprintf("%d poin · %d sumber", len(items), len(sources)),
			Items:    items, Sources: sources, Muted: muted,
		}
	}
	eraporSrc := "Data siswa"
	if top != nil {
		eraporSrc = "E-Rapor · " + eraporDate
	}
	insights := []Insight{
		mkInsight("idea", "Kemampuan & minat", abilities, []string{eraporSrc, "Minat & bakat"}, false),
		mkInsight("star", "Kelebihan", strengths, []string{"E-Rapor · " + eraporDate, "Pendapat wali kelas"}, false),
		mkInsight("favourite", "Kebutuhan & dukungan", needs, []string{obsSource(in, lastObsDate)}, false),
		mkInsight("infoCircle", "Data belum cukup", notEnough, []string{"Catatan kasus · Kosong", "Catatan psikologis · Tidak dikirim"}, true),
	}

	// --- Data used: 8 baris ---
	dataUsed := []DataUsedRow{
		{Source: "Data siswa", Details: "Nama, kelas, NISN", Updated: "Hari ini", Status: "used"},
		{Source: "Nilai E-Rapor", Details: fmt.Sprintf("%d mapel", len(scores)), Updated: eraporDate, Status: usedMissing(len(scores) > 0)},
		{Source: "Observasi", Details: obsDetails(in), Updated: obsUpdated(in, lastObsDate), Status: usedMissing(in.ObsCount > 0)},
		{Source: "Pendapat wali kelas", Details: filled(authorTitle, in.PendapatWaliKelas), Updated: filledDate(in.PendapatWaliKelas), Status: usedMissing(nonEmpty(in.PendapatWaliKelas))},
		{Source: "Masukan BK", Details: filled("Sudah diisi", in.MasukanBK), Updated: filledDate(in.MasukanBK), Status: usedMissing(nonEmpty(in.MasukanBK))},
		{Source: "Minat & bakat", Details: filled(authorTitle, in.MinatBakat), Updated: filledDate(in.MinatBakat), Status: usedMissing(nonEmpty(in.MinatBakat))},
		{Source: "Catatan kasus & prestasi", Details: filled("Tersedia", in.CatatanKasusPrestasi), Updated: filledDate(in.CatatanKasusPrestasi), Status: usedMissing(nonEmpty(in.CatatanKasusPrestasi))},
		{Source: "Catatan psikologis", Details: "Akses terbatas", Updated: "—", Status: "not_sent"},
	}

	var rf *RedFlagNote
	if hasRed {
		title := in.RedFlagTitle
		if !nonEmpty(title) {
			title = "Red flag terbuka · BK sudah diberi tahu"
		}
		rf = &RedFlagNote{
			Title: title,
			Note:  "AI tidak memberi skor atau menilai risiko. Tindak lanjut tetap di tangan BK.",
		}
	}

	return SummaryContent{
		Headline:  headline,
		Narrative: narrative,
		Metrics: []Metric{
			{Label: metricTopLabel(top, avg), Value: metricTopValue(top, avg)},
			{Label: fmt.Sprintf("Observasi · %d red flag", in.RedFlagCount), Value: fmt.Sprintf("%d", in.ObsCount)},
			{Label: "Kategori indikator", Value: fmt.Sprintf("%d", len(in.Categories))},
			{Label: "Sumber data belum ada", Value: fmt.Sprintf("%d", 8-countUsed(dataUsed))},
		},
		Insights: insights,
		DataUsed: dataUsed,
		RedFlag:  rf,
	}
}

func countUsed(rows []DataUsedRow) int {
	n := 0
	for _, r := range rows {
		if r.Status == "used" {
			n++
		}
	}
	return n
}

func obsSource(in ContentInput, lastObsDate string) string {
	if in.ObsCount > 0 {
		return "Observasi · " + lastObsDate
	}
	return "Observasi"
}

func obsDetails(in ContentInput) string {
	if in.ObsCount == 0 {
		return "0 catatan"
	}
	if in.RedFlagCount > 0 {
		return fmt.Sprintf("%d catatan · %d red flag", in.ObsCount, in.RedFlagCount)
	}
	return fmt.Sprintf("%d catatan", in.ObsCount)
}

func obsUpdated(in ContentInput, lastObsDate string) string {
	if in.ObsCount > 0 {
		return lastObsDate
	}
	return "—"
}

func usedMissing(ok bool) string {
	if ok {
		return "used"
	}
	return "missing"
}

func filled(v, s string) string {
	if nonEmpty(s) {
		return v
	}
	return "Belum diisi"
}

func filledDate(s string) string {
	if nonEmpty(s) {
		return "Hari ini"
	}
	return "—"
}

func metricTopLabel(top *EraporRow, avg int) string {
	if top != nil {
		return top.Mapel + " · nilai tertinggi"
	}
	_ = avg
	return "Rata-rata nilai"
}

func metricTopValue(top *EraporRow, avg int) string {
	if top != nil {
		return trimNum(top.Nilai)
	}
	return fmt.Sprintf("%d", avg)
}

// trimNum: 88 → "88"; 87.5 → "87.5".
func trimNum(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", f), "0"), ".")
}
