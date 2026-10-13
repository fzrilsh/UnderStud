package service

import (
	"sort"
	"strings"

	"backend/internal/utils"
)

// ChecklistItem = satu item observasi (referensi tetap, spec §4.4).
type ChecklistItem struct {
	Key      string `json:"key"`      // "<kategori>:<item>"
	Kategori string `json:"kategori"`
	Label    string `json:"label"`
}

// checklistCatalog = 28 item + tidak_ada, hardcoded (spec §4.4).
var checklistCatalog = []ChecklistItem{
	// 1 — Emosi
	{"emosi:mudah_marah", "emosi", "Mudah marah atau emosi meledak"},
	{"emosi:mudah_menangis", "emosi", "Mudah menangis tanpa sebab jelas"},
	{"emosi:murung", "emosi", "Sering murung atau terlihat sedih"},
	{"emosi:mood_berubah", "emosi", "Mood berubah drastis dalam sehari"},
	{"emosi:cemas_berlebihan", "emosi", "Tampak cemas berlebihan (gelisah, takut ditanya, takut ujian)"},
	// 2 — Sosial
	{"sosial:menyendiri", "sosial", "Menyendiri atau menarik diri dari teman"},
	{"sosial:konflik", "sosial", "Sering konflik atau berkelahi dengan teman"},
	{"sosial:korban_ejekan", "sosial", "Jadi korban ejekan atau dijauhi"},
	{"sosial:mendominasi", "sosial", "Mengejek atau mendominasi teman lain"},
	{"sosial:ganti_pertemanan", "sosial", "Tiba-tiba ganti lingkaran pertemanan"},
	// 3 — Akademik dan motivasi
	{"akademik:nilai_turun", "akademik", "Nilai atau kualitas tugas turun mendadak"},
	{"akademik:sulit_fokus", "akademik", "Sulit fokus atau sering melamun di kelas"},
	{"akademik:hilang_minat", "akademik", "Kehilangan minat pada pelajaran/kegiatan yang dulu disukai"},
	{"akademik:tidak_kerjakan_tugas", "akademik", "Sering tidak mengerjakan tugas"},
	{"akademik:perfeksionis", "akademik", "Terlalu perfeksionis atau panik kalau salah"},
	// 4 — Kehadiran dan disiplin
	{"kehadiran:terlambat_bolos", "kehadiran", "Sering terlambat atau bolos"},
	{"kehadiran:izin_sakit", "kehadiran", "Sering izin sakit tanpa kejelasan"},
	{"kehadiran:melanggar_aturan", "kehadiran", "Melanggar aturan berulang"},
	// 5 — Fisik dan kebiasaan
	{"fisik:mengantuk_lelah", "fisik", "Sering mengantuk atau kelelahan"},
	{"fisik:nafsu_makan", "fisik", "Perubahan nafsu makan atau berat badan yang terlihat"},
	{"fisik:sakit_fisik", "fisik", "Sering mengeluh sakit kepala atau perut"},
	{"fisik:kebersihan_menurun", "fisik", "Penampilan atau kebersihan diri menurun drastis"},
	// 6 — Situasi keluarga
	{"keluarga:perubahan_kondisi", "keluarga", "Perubahan kondisi di rumah (perceraian, kehilangan, masalah ekonomi)"},
	{"keluarga:kurang_diperhatikan", "keluarga", "Tidak ada yang menjemput atau tampak kurang diperhatikan di rumah"},
	// 7 — Red flag
	{"red_flag:menyakiti_diri", "red_flag", "Mengucapkan ingin menyakiti diri atau tidak ingin hidup"},
	{"red_flag:luka_tidak_wajar", "red_flag", "Luka atau memar yang tidak wajar"},
	{"red_flag:mengancam", "red_flag", "Mengancam orang lain"},
	{"red_flag:takut_orang", "red_flag", "Tampak ketakutan terhadap seseorang tertentu"},
}

var checklistSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(checklistCatalog))
	for _, it := range checklistCatalog {
		m[it.Key] = struct{}{}
	}
	return m
}()

// ChecklistCatalog mengembalikan salinan catalog (untuk endpoint FE).
func ChecklistCatalog() []ChecklistItem {
	out := make([]ChecklistItem, len(checklistCatalog))
	copy(out, checklistCatalog)
	return out
}

// ValidateItems memvalidasi item mentah, mengembalikan items ter-dedup,
// categories (turunan, urut), dan redFlag (spec §4.4).
func ValidateItems(raw []string) ([]string, []string, bool, error) {
	if len(raw) == 0 {
		return nil, nil, false, utils.NewErrInvalidInput("Pilih minimal satu item observasi.")
	}
	seen := map[string]struct{}{}
	var items []string
	hasTidakAda := false
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if r == "tidak_ada" {
			hasTidakAda = true
			continue
		}
		if _, ok := checklistSet[r]; !ok {
			return nil, nil, false, utils.NewErrInvalidInput("Item observasi tidak dikenal: " + r)
		}
		if _, dup := seen[r]; dup {
			continue
		}
		seen[r] = struct{}{}
		items = append(items, r)
	}
	if hasTidakAda && len(items) > 0 {
		return nil, nil, false, utils.NewErrInvalidInput("'Tidak ada indikator' tidak bisa digabung dengan item lain.")
	}
	if hasTidakAda {
		return []string{"tidak_ada"}, nil, false, nil
	}
	if len(items) == 0 {
		return nil, nil, false, utils.NewErrInvalidInput("Pilih minimal satu item observasi.")
	}
	catSet := map[string]struct{}{}
	redFlag := false
	for _, it := range items {
		cat := it[:strings.IndexByte(it, ':')]
		catSet[cat] = struct{}{}
		if cat == "red_flag" {
			redFlag = true
		}
	}
	cats := make([]string, 0, len(catSet))
	for c := range catSet {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return items, cats, redFlag, nil
}
