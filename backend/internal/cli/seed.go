package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/service"
	"backend/internal/utils"
)

// Kelas SMA: hanya IPA & IPS (tidak ada TJKT/teknik).
var seedKelas = []string{
	"X IPA 1", "X IPA 2", "X IPS 1",
	"XI IPA 1", "XI IPA 2", "XI IPS 1",
	"XII IPA 1", "XII IPS 1",
}

// Wali kelas per kelas (staff demo).
var seedWali = []struct {
	email, name, kelas string
}{
	{"wali@demo.local", "Budi Santoso", "X IPA 1"},
	{"wali.ipa2@demo.local", "Siti Nurhaliza", "X IPA 2"},
	{"wali.ips1@demo.local", "Agus Wijaya", "X IPS 1"},
	{"wali.xi.ipa1@demo.local", "Dewi Lestari", "XI IPA 1"},
	{"wali.xi.ipa2@demo.local", "Hendra Gunawan", "XI IPA 2"},
	{"wali.xi.ips1@demo.local", "Ratna Sari", "XI IPS 1"},
	{"wali.xii.ipa1@demo.local", "Joko Susilo", "XII IPA 1"},
	{"wali.xii.ips1@demo.local", "Maya Anggraini", "XII IPS 1"},
}

// Item checklist valid (28) — dipakai generate observasi acak.
var (
	obsEmosi     = []string{"emosi:mudah_marah", "emosi:mudah_menangis", "emosi:murung", "emosi:mood_berubah", "emosi:cemas_berlebihan"}
	obsSosial    = []string{"sosial:menyendiri", "sosial:konflik", "sosial:korban_ejekan", "sosial:mendominasi", "sosial:ganti_pertemanan"}
	obsAkademik  = []string{"akademik:nilai_turun", "akademik:sulit_fokus", "akademik:hilang_minat", "akademik:tidak_kerjakan_tugas", "akademik:perfeksionis"}
	obsKehadiran = []string{"kehadiran:terlambat_bolos", "kehadiran:izin_sakit", "kehadiran:melanggar_aturan"}
	obsFisik     = []string{"fisik:mengantuk_lelah", "fisik:nafsu_makan", "fisik:sakit_fisik", "fisik:kebersihan_menurun"}
	obsKeluarga  = []string{"keluarga:perubahan_kondisi", "keluarga:kurang_diperhatikan"}
	obsRedFlag   = []string{"red_flag:menyakiti_diri", "red_flag:luka_tidak_wajar", "red_flag:mengancam", "red_flag:takut_orang"}
	allObs       = [][]string{obsEmosi, obsSosial, obsAkademik, obsKehadiran, obsFisik, obsKeluarga}
)

// Nama siswa Indonesia (depan + belakang) untuk generate banyak siswa.
var (
	namaDepan = []string{
		"Aditya", "Bunga", "Citra", "Dimas", "Eka", "Fajar", "Gita", "Hendra", "Indah", "Joko",
		"Kartika", "Lukman", "Maya", "Nadia", "Oscar", "Putri", "Rizki", "Sari", "Taufik", "Umi",
		"Vina", "Wahyu", "Yuni", "Zaki", "Anisa", "Bayu", "Cahya", "Dewi", "Eko", "Fitri",
		"Galih", "Hana", "Irfan", "Julia", "Kevin", "Lina", "Mega", "Nanda", "Omar", "Prita",
	}
	namaBelakang = []string{
		"Pratama", "Lestari", "Dewi", "Saputra", "Wijaya", "Santoso", "Nurhaliza", "Gunawan",
		"Anggraini", "Susilo", "Ramadhan", "Maharani", "Kusuma", "Hidayat", "Permata", "Sari",
		"Firmansyah", "Utami", "Setiawan", "Rahmawati",
	}
	minatList = []string{
		"Basket", "Sepak bola", "Menulis", "Menggambar", "Musik", "Coding", "Debat", "Fotografi",
		"Menari", "Bulu tangkis", "Renang", "Membaca", "Voli", "Teater", "Melukis", "Berkebun",
	}
	psikList = []string{
		"",
		"",
		"",
		"Sedang menghadapi perceraian orang tua.",
		"Cemas menjelang ujian besar.",
		"Kurang percaya diri saat presentasi.",
		"Perlu dukungan adaptasi setelah pindah sekolah.",
		"Tekanan dari ekspektasi orang tua.",
	}
	catatanKasusList = []string{
		"",
		"",
		"Juara 2 Olimpiade Sains tingkat kabupaten.",
		"Anggota OSIS bidang sosial.",
		"Ketua kelas semester lalu.",
		"Pernah terlibat perkelahian kecil (sudah diselesaikan).",
	}
	obsCatatan = []string{
		"Terlihat murung 2 minggu terakhir.",
		"Nilai matematika turun drastis.",
		"Bolos 3x minggu ini.",
		"Sering menyendiri di kantin.",
		"Tidur di kelas beberapa kali.",
		"Aktif bertanya, tapi tampak cemas.",
		"Sering izin dengan alasan sakit.",
		"Tidak mengumpulkan tugas 2 minggu berturut-turut.",
	}
	tindakanList = []string{"pantau", "konsultasi_bk", "eskalasi_segera"}
)

// mapelIPA/mapelIPS: mapel khas jurusan (SMA).
func mapelIPA() []string {
	return []string{"Matematika (Umum)", "Fisika", "Kimia", "Biologi", "Bahasa Indonesia", "Bahasa Inggris"}
}
func mapelIPS() []string {
	return []string{"Matematika (Umum)", "Ekonomi", "Sosiologi", "Geografi", "Sejarah Indonesia", "Bahasa Indonesia"}
}

// SeedDemo mengisi data demo lengkap (idempotent — aman dijalankan berulang).
// Konteks: SMA (jurusan IPA & IPS), 8 kelas, ~8 siswa/kelas + observasi,
// profil bervariasi, ringkasan tidak di-generate (butuh AI), dan chat anonim.
func SeedDemo(cfg config.Config) error {
	pool, q, err := connect(cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	ctx := context.Background()
	rng := rand.New(rand.NewSource(42)) // deterministik antar-run

	if err := upsertStaff(ctx, q, "bk@demo.local", "password123", "Rina Pratiwi", "BK", nil); err != nil {
		return err
	}
	// Beberapa Guru BK tambahan (biar daftar staff terasa nyata).
	if err := upsertStaff(ctx, q, "bk2@demo.local", "password123", "Ahmad Fauzi", "BK", nil); err != nil {
		return err
	}

	// Wali kelas: upsert semua.
	for _, w := range seedWali {
		kelas := w.kelas
		if err := upsertStaff(ctx, q, w.email, "password123", w.name, "WALI_KELAS", &kelas); err != nil {
			return err
		}
	}

	// Siswa per kelas.
	nisnCounter := 2060691500
	for _, kelas := range seedKelas {
		// Cari wali kelas ini.
		var waliEmail string
		for _, w := range seedWali {
			if w.kelas == kelas {
				waliEmail = w.email
				break
			}
		}
		wali, err := q.GetUserByEmail(ctx, waliEmail)
		if err != nil {
			return err
		}

		isIPA := len(kelas) >= 6 && kelas[2:5] == "IPA"
		mapel := mapelIPS()
		if isIPA {
			mapel = mapelIPA()
		}

		jumlah := 8
		for i := 0; i < jumlah; i++ {
			nisnCounter++
			nisn := fmt.Sprintf("%d", nisnCounter)
			nama := namaDepan[rng.Intn(len(namaDepan))] + " " + namaBelakang[rng.Intn(len(namaBelakang))]

			st, err := q.GetStudentByNISN(ctx, &nisn)
			if err != nil {
				st, err = q.CreateStudent(ctx, db.CreateStudentParams{Nisn: &nisn, Name: nama, Kelas: kelas})
				if err != nil {
					return err
				}
				if err := q.CreateStudentProfileEmpty(ctx, st.ID); err != nil {
					return err
				}
			}

			// Profil: nilai bervariasi (75–95), kehadiran, minat, psikologis.
			nilai := make([]map[string]any, 0, len(mapel))
			for _, m := range mapel {
				nilai = append(nilai, map[string]any{
					"mapel": m, "nilai": 75 + rng.Intn(21), "periode": "2025/2026 Genap",
				})
			}
			minat := minatList[rng.Intn(len(minatList))]
			psik := psikList[rng.Intn(len(psikList))]
			kasus := catatanKasusList[rng.Intn(len(catatanKasusList))]
			sakit, izin, alpa := rng.Intn(4), rng.Intn(4), rng.Intn(3)
			if err := seedProfile(ctx, q, st.ID, wali.ID, nilai, minat, psik, kasus, sakit, izin, alpa); err != nil {
				return err
			}

			// Observasi: 0–3 entri per siswa.
			existing, err := q.CountObservationsByStudent(ctx, st.ID)
			if err != nil {
				return err
			}
			if existing > 0 {
				continue
			}
			nObs := rng.Intn(4) // 0..3
			for o := 0; o < nObs; o++ {
				items := randomObsItems(rng)
				items2, cats, redFlag, err := service.ValidateItems(items)
				if err != nil {
					return err
				}
				itemsJSON, _ := json.Marshal(items2)
				catsJSON, _ := json.Marshal(cats)
				catatan := obsCatatan[rng.Intn(len(obsCatatan))]
				tindakan := tindakanList[rng.Intn(len(tindakanList))]
				if redFlag {
					tindakan = "eskalasi_segera"
				}
				tanggal := time.Now().AddDate(0, 0, -(rng.Intn(60) + 1))
				if _, err := q.CreateObservation(ctx, db.CreateObservationParams{
					StudentID: st.ID, AuthorID: wali.ID,
					Items: itemsJSON, Categories: catsJSON, RedFlag: redFlag,
					Catatan: &catatan, Tanggal: pgtype.Date{Time: tanggal, Valid: true},
					TindakanAwal: &tindakan,
				}); err != nil {
					return err
				}
			}
		}
	}

	// Chat anonim: 12 contoh dengan variasi status & jumlah pesan.
	if err := seedAnonChats(ctx, q); err != nil {
		return err
	}

	fmt.Printf("OK — seed demo SMA selesai (%d kelas, akun: bk@demo.local / wali@demo.local, password: password123)\n", len(seedKelas))
	return nil
}

// randomObsItems: 1–3 kategori acak, 1–2 item per kategori; red flag jarang (~12%).
func randomObsItems(rng *rand.Rand) []string {
	var items []string
	nKat := rng.Intn(3) + 1
	for i := 0; i < nKat; i++ {
		pool := allObs[rng.Intn(len(allObs))]
		n := rng.Intn(2) + 1
		for j := 0; j < n; j++ {
			items = append(items, pool[rng.Intn(len(pool))])
		}
	}
	// Red flag muncul jarang, dan selalu bersama indikator lain.
	if rng.Intn(100) < 12 {
		items = append(items, obsRedFlag[rng.Intn(len(obsRedFlag))])
	}
	return items
}

// seedAnonChats membuat 12 chat anonim contoh (idempotent via token hash tetap).
func seedAnonChats(ctx context.Context, q *db.Queries) error {
	type chatSeed struct {
		body     string
		balasan  string // "" = belum dibalas
		ident    string // anonim | profil
		nama     string
		kelas    string
	}
	chats := []chatSeed{
		{"Saya sering merasa cemas saat ujian. Boleh minta saran?", "Terima kasih sudah bercerita. Kapan kamu merasa paling cemas?", "anonim", "", ""},
		{"Akhir-akhir ini saya sulit fokus belajar, sering melamun di kelas.", "Sejak kapan ini terasa mengganggu belajarmu?", "anonim", "", ""},
		{"Saya merasa sendirian di kelas, teman-teman menjauh dari saya.", "Terima kasih sudah bercerita. Boleh ceritakan lebih detail kapan ini mulai terjadi?", "anonim", "", ""},
		{"Orang tua saya bertengkar terus di rumah, saya jadi tidak nyaman belajar.", "", "anonim", "", ""},
		{"Saya ingin menceritakan sesuatu tapi takut diketahui teman.", "Kamu bisa memakai mode anonim di sini. Ceritakan saja, kami mendengarkan.", "anonim", "", ""},
		{"Nilai saya turun terus semester ini, padahal saya sudah belajar.", "Mari kita lihat bersama. Mapel mana yang paling menurun?", "anonim", "", ""},
		{"Saya sering sakit perut setiap pagi sebelum sekolah.", "Apakah ini sering terjadi saat hari tertentu? Bagaimana perasaanmu saat itu?", "anonim", "", ""},
		{"Teman sekelas saya sering mengejek bentuk tubuh saya.", "Terima kasih sudah berani melapor. Ini tidak boleh dibiarkan — kami akan menindaklanjuti.", "profil", "Rahma", "X IPA 2"},
		{"Saya bingung mau pilih jurusan kuliah apa. Orang tua ingin saya kedokteran tapi saya suka seni.", "Itu hal yang wajar. Mau kita diskusikan minat dan kekuatanmu?", "profil", "Dimas", "XI IPS 1"},
		{"Kakak saya sering memukul saya kalau salah.", "Kamu aman di sini. Kami akan bantu, tapi tolong ceritakan kapan biasanya terjadi.", "anonim", "", ""},
		{"Saya takut presentasi di depan kelas, sampai tidak masuk sekolah beberapa kali.", "Kita bisa latih bersama sedikit-sedikit. Kapan presentasi berikutnya?", "anonim", "", ""},
		{"Saya ingin berhenti sekolah dan kerja membantu keluarga.", "", "anonim", "", ""},
	}

	for i, c := range chats {
		tokenHash := fmt.Sprintf("seed-demo-anon-chat-%02d-fixed-0000000000000000000000000000000000000000", i)
		sess, err := q.GetAnonSessionByTokenHash(ctx, tokenHash)
		if err != nil {
			var name, kelas *string
			if c.nama != "" {
				name, kelas = &c.nama, &c.kelas
			}
			sess, err = q.CreateAnonSession(ctx, tokenHash)
			if err != nil {
				return err
			}
			if name != nil {
				if _, err := q.UpdateAnonProfile(ctx, db.UpdateAnonProfileParams{
					ID: sess.ID, DisplayName: name, DisplayKelas: kelas,
				}); err != nil {
					return err
				}
			}
		}
		n, err := q.CountChatsBySession(ctx, sess.ID)
		if err != nil {
			return err
		}
		if n > 0 {
			continue // sudah ada
		}
		var snapName, snapKelas *string
		if c.ident == "profil" {
			snapName, snapKelas = &c.nama, &c.kelas
		}
		chat, err := q.CreateChat(ctx, db.CreateChatParams{
			SessionID: sess.ID, IdentityMode: c.ident, SnapshotName: snapName, SnapshotKelas: snapKelas,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateMessage(ctx, db.CreateMessageParams{
			ChatID: chat.ID, Sender: "SISWA", SenderUserID: nil, Body: c.body,
		}); err != nil {
			return err
		}
		if c.balasan != "" {
			if _, err := q.CreateMessage(ctx, db.CreateMessageParams{
				ChatID: chat.ID, Sender: "BK", SenderUserID: nil, Body: c.balasan,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func seedProfile(ctx context.Context, q *db.Queries, studentID, waliID uuid.UUID, nilai []map[string]any, minat, psik, kasus string, sakit, izin, alpa int) error {
	nilaiJSON, _ := json.Marshal(nilai)
	params := db.UpsertStudentProfileParams{
		StudentID: studentID, NilaiErapor: nilaiJSON, UpdatedBy: &waliID,
		Sakit: int32(sakit), Izin: int32(izin), Alpa: int32(alpa),
	}
	if minat != "" {
		params.MinatBakat = &minat
	}
	if psik != "" {
		params.CatatanPsikologis = &psik
	}
	if kasus != "" {
		params.CatatanKasusPrestasi = &kasus
	}
	_, err := q.UpsertStudentProfile(ctx, params)
	return err
}

func upsertStaff(ctx context.Context, q *db.Queries, email, pass, name, role string, kelas *string) error {
	existing, err := q.GetUserByEmail(ctx, email)
	if err == nil {
		// Sudah ada → sinkronkan role/kelas (mis. kelas seed berubah antar versi).
		_, err = q.UpdateStaff(ctx, db.UpdateStaffParams{
			Email: existing.Email, Name: name, Role: role, Kelas: kelas,
		})
		return err
	}
	hash, err := utils.HashPassword(pass)
	if err != nil {
		return err
	}
	_, err = q.CreateStaff(ctx, db.CreateStaffParams{
		Email: email, PasswordHash: hash, Name: name, Role: role, Kelas: kelas,
	})
	return err
}
