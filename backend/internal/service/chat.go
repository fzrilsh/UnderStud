package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/geo"
	"backend/internal/middleware"
	"backend/internal/storage"
	"backend/internal/utils"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrOutOfScope = errors.New("out of scope")
	ErrChatHidden = errors.New("chat hidden")
)

type Chat struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	store   *storage.Store
	rl      *middleware.RateLimiter
	cfg     config.Config
}

func NewChat(q *db.Queries, pool *pgxpool.Pool, store *storage.Store, rl *middleware.RateLimiter, cfg config.Config) *Chat {
	return &Chat{queries: q, pool: pool, store: store, rl: rl, cfg: cfg}
}

type UploadFile struct {
	Name string
	Data []byte
}

type CreateChatInput struct {
	Body         string
	IdentityMode string
	Files        []UploadFile
	Lat, Lng     float64
	Accuracy     float64
	ClientIP     string // untuk rate limit per-IP (audit M4)
}

type ReplyInput struct {
	Body     string
	Files    []UploadFile
	ClientIP string
}

type MessageView struct {
	ID          string           `json:"id"`
	Sender      string           `json:"sender"`
	Body        string           `json:"body"`
	Attachments []AttachmentView `json:"attachments"`
	CreatedAt   string           `json:"created_at"`
}

type AttachmentView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mime string `json:"mime"`
	Size int32  `json:"size_bytes"`
}

type ChatSummary struct {
	ID            string       `json:"id"`
	IdentityMode  string       `json:"identity_mode"`
	Status        string       `json:"status"`
	LastMessageAt string       `json:"last_message_at"`
	LastMessage   *LastMessage `json:"last_message"`  // preview pesan terakhir
	HasNewReply   bool         `json:"has_new_reply"` // F23: pesan terakhir dari BK
	UnreadCount   int32        `json:"unread_count"`  // balasan BK belum dibaca siswa
}

type ChatView struct {
	ID           string `json:"id"`
	IdentityMode string `json:"identity_mode"`
	Status       string `json:"status"`
}

type ChatDetail struct {
	Chat     ChatView      `json:"chat"`
	Messages []MessageView `json:"messages"`
}

func (c *Chat) checkGeo(lat, lng, accuracy float64) error {
	if !c.cfg.GeoFenceEnabled {
		return nil
	}
	// M6: tolak koordinat non-finite (NaN/Inf) — kalau tidak, cek rentang lolos
	// (semua perbandingan NaN = false) sehingga geofence bisa di-bypass.
	if !geo.ValidCoord(lat, lng) {
		return utils.NewErrLocationUnavailable("Aktifkan izin lokasi untuk mengirim laporan.")
	}
	if lat == 0 && lng == 0 {
		return utils.NewErrLocationUnavailable("Aktifkan izin lokasi untuk mengirim laporan.")
	}
	dist := geo.DistanceM(lat, lng, c.cfg.SchoolLat, c.cfg.SchoolLng)
	if dist > float64(c.cfg.SchoolRadiusM) {
		return utils.NewErrOutOfRange("Kamu di luar area sekolah. Laporan hanya bisa dikirim dari lingkungan sekolah.")
	}
	// F18: accuracy NaN lolos cek (NaN > radius = false) — perlakukan sebagai
	// tidak valid (minta aktifkan ulang izin lokasi).
	if math.IsNaN(accuracy) || math.IsInf(accuracy, 0) || accuracy > float64(c.cfg.SchoolRadiusM) {
		return utils.NewErrLocationImprecise("Sinyal lokasi kurang akurat. Coba lagi di area lebih terbuka.")
	}
	return nil
}

func (c *Chat) Create(ctx context.Context, session db.AnonSession, in CreateChatInput) (ChatView, error) {
	body := strings.TrimSpace(in.Body)
	if len([]rune(body)) > 2000 {
		return ChatView{}, utils.NewErrInvalidInput("Laporan maksimal 2.000 karakter.")
	}
	if body == "" && len(in.Files) == 0 {
		return ChatView{}, utils.NewErrInvalidInput("Tulis laporan sebelum mengirim.")
	}
	if err := c.rl.CheckChat(session.ID.String(), in.ClientIP); err != nil {
		return ChatView{}, err
	}
	if err := c.checkGeo(in.Lat, in.Lng, in.Accuracy); err != nil {
		return ChatView{}, err
	}
	mode := in.IdentityMode
	if mode != "profil" {
		mode = "anonim"
	}
	var snapName, snapKelas *string
	if mode == "profil" {
		if session.DisplayName == nil || session.DisplayKelas == nil ||
			*session.DisplayName == "" || *session.DisplayKelas == "" {
			return ChatView{}, utils.NewErrInvalidInput("Lengkapi nama dan kelas dulu, atau pilih mode Anonim.")
		}
		snapName, snapKelas = session.DisplayName, session.DisplayKelas
	}

	saved, err := c.saveFiles(in.Files)
	if err != nil {
		return ChatView{}, err
	}

	// Audit lokasi (spec §11.4 poin 6): catat hanya bila koordinat valid
	// (NaN/Inf/rentang salah → NULL, jangan simpan sampah).
	var geoLat, geoLng, geoAcc *float64
	if geo.ValidCoord(in.Lat, in.Lng) && (in.Lat != 0 || in.Lng != 0) {
		geoLat, geoLng = &in.Lat, &in.Lng
		if !math.IsNaN(in.Accuracy) && !math.IsInf(in.Accuracy, 0) && in.Accuracy > 0 {
			geoAcc = &in.Accuracy
		}
	}

	// F8: chat + pesan pertama + lampiran dalam SATU transaksi. Bila ada langkah
	// gagal → rollback DB dan hapus file tersimpan (cegah chat kosong / file orphan).
	var view ChatView
	err = db.WithTx(ctx, c.pool, func(q *db.Queries) error {
		chat, err := q.CreateChat(ctx, db.CreateChatParams{
			SessionID:     session.ID,
			IdentityMode:  mode,
			SnapshotName:  snapName,
			SnapshotKelas: snapKelas,
			GeoLat:        geoLat,
			GeoLng:        geoLng,
			GeoAccuracy:   geoAcc,
		})
		if err != nil {
			return utils.NewErrInternal("gagal membuat laporan")
		}
		msg, err := q.CreateMessage(ctx, db.CreateMessageParams{
			ChatID: chat.ID, Sender: "SISWA", SenderUserID: nil, Body: body,
		})
		if err != nil {
			return utils.NewErrInternal("gagal menyimpan pesan")
		}
		if err := attachFilesTx(ctx, q, msg.ID, saved); err != nil {
			return err
		}
		view = ChatView{ID: chat.ID.String(), IdentityMode: chat.IdentityMode, Status: chat.Status}
		return nil
	})
	if err != nil {
		c.discardFiles(saved)
		return ChatView{}, err
	}
	return view, nil
}

// saveFiles memvalidasi & menyimpan; mengembalikan daftar file tersimpan.
func (c *Chat) saveFiles(files []UploadFile) ([]storage.SavedFile, error) {
	// M10: >3 file = payload terlalu besar (413), bukan INVALID_INPUT.
	if len(files) > 3 {
		return nil, utils.NewErrPayloadTooLarge("Maksimal 3 file per pesan.")
	}
	// M10: cek total SEBELUM menulis ke disk — cegah file orphan saat total
	// melebihi batas (input bytes = batas atas konservatif).
	maxTotal := c.cfg.MaxUploadTotalMB * 1024 * 1024
	preTotal := 0
	for _, f := range files {
		preTotal += len(f.Data)
	}
	if preTotal > maxTotal {
		return nil, utils.NewErrPayloadTooLarge("Total lampiran maksimal 20 MB per pesan.")
	}

	total := 0
	var out []storage.SavedFile
	for _, f := range files {
		sf, err := c.store.SaveOne(f.Data, f.Name)
		if err != nil {
			if errors.Is(err, storage.ErrUnsupportedMedia) {
				return nil, utils.NewErrUnsupportedMedia("Hanya gambar (JPG/PNG/WebP/GIF), PDF, atau TXT.")
			}
			if errors.Is(err, storage.ErrTooLarge) {
				return nil, utils.NewErrPayloadTooLarge("File maksimal 10 MB.")
			}
			return nil, utils.NewErrInternal("Gagal menyimpan file. Coba lagi.")
		}
		total += sf.Size
		if total > maxTotal {
			return nil, utils.NewErrPayloadTooLarge("Total lampiran maksimal 20 MB per pesan.")
		}
		out = append(out, sf)
	}
	return out, nil
}

func attachFilesTx(ctx context.Context, q *db.Queries, messageID uuid.UUID, saved []storage.SavedFile) error {
	for _, sf := range saved {
		name := sf.OriginalName
		if _, err := q.CreateAttachment(ctx, db.CreateAttachmentParams{
			MessageID:    messageID,
			StoredPath:   sf.StoredPath,
			OriginalName: &name,
			Mime:         sf.Mime,
			SizeBytes:    int32(sf.Size),
		}); err != nil {
			return utils.NewErrInternal("gagal menyimpan lampiran")
		}
	}
	return nil
}

// discardFiles menghapus file yang sudah ditulis ke disk saat transaksi DB
// gagal (F8) — cegah file orphan.
func (c *Chat) discardFiles(saved []storage.SavedFile) {
	for _, sf := range saved {
		c.store.Delete(sf.StoredPath)
	}
}

// fillAttachments mengisi lampiran untuk sekumpulan pesan dalam SATU query
// (F4: hindari N+1). Dipakai Chat & Report.
func fillAttachments(ctx context.Context, q *db.Queries, views []MessageView, msgs []db.Message) ([]MessageView, error) {
	if len(msgs) == 0 {
		return views, nil
	}
	ids := make([]uuid.UUID, 0, len(msgs))
	for i := range msgs {
		ids = append(ids, msgs[i].ID)
	}
	atts, err := q.ListAttachmentsByMessages(ctx, ids)
	if err != nil {
		return nil, utils.NewErrInternal("gagal memuat lampiran")
	}
	if len(atts) == 0 {
		return views, nil
	}
	byID := make(map[string]int, len(msgs))
	for i, m := range msgs {
		byID[m.ID.String()] = i
	}
	for _, a := range atts {
		idx, ok := byID[a.MessageID.String()]
		if !ok {
			continue
		}
		name := "lampiran"
		if a.OriginalName != nil && *a.OriginalName != "" {
			name = *a.OriginalName
		}
		views[idx].Attachments = append(views[idx].Attachments, AttachmentView{
			ID: a.ID.String(), Name: name, Mime: a.Mime, Size: a.SizeBytes,
		})
	}
	return views, nil
}

func (c *Chat) ListBySession(ctx context.Context, session db.AnonSession, limit, offset int32) ([]ChatSummary, int64, error) {
	chats, err := c.queries.ListChatsBySession(ctx, db.ListChatsBySessionParams{
		SessionID: session.ID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat daftar")
	}
	out := make([]ChatSummary, 0, len(chats))
	if len(chats) > 0 {
		ids := make([]uuid.UUID, 0, len(chats))
		for _, ch := range chats {
			ids = append(ids, ch.ID)
		}
		lastByChat := map[uuid.UUID]LastMessage{}
		if rows, err := c.queries.ListLastMessages(ctx, ids); err == nil {
			for _, r := range rows {
				lastByChat[r.ChatID] = LastMessage{
					Sender: r.Sender, Body: r.Body, CreatedAt: r.CreatedAt.Format(time.RFC3339),
				}
			}
		}
		for _, ch := range chats {
			sum := ChatSummary{
				ID: ch.ID.String(), IdentityMode: ch.IdentityMode, Status: ch.Status,
				LastMessageAt: ch.LastMessageAt.Format(time.RFC3339),
				HasNewReply:   lastByChat[ch.ID].Sender == "BK",
				UnreadCount:   ch.StudentUnreadCount,
			}
			if lm, ok := lastByChat[ch.ID]; ok {
				sum.LastMessage = &lm
			}
			out = append(out, sum)
		}
	}
	// H5: total = jumlah seluruh baris (bukan ukuran halaman).
	total, err := c.queries.CountChatsBySession(ctx, session.ID)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat daftar")
	}
	return out, total, nil
}

// ownedChat memastikan chat milik sesi ini (cegah akses lintas sesi).
func (c *Chat) ownedChat(ctx context.Context, session db.AnonSession, chatID uuid.UUID) (db.Chat, error) {
	chat, err := c.queries.GetChatByID(ctx, chatID)
	if err != nil {
		return db.Chat{}, utils.NewErrNotFound("Laporan tidak ditemukan.")
	}
	if chat.SessionID != session.ID {
		return db.Chat{}, utils.NewErrNotFound("Laporan tidak ditemukan.")
	}
	return chat, nil
}

func (c *Chat) Detail(ctx context.Context, session db.AnonSession, chatID uuid.UUID) (ChatDetail, error) {
	chat, err := c.ownedChat(ctx, session, chatID)
	if err != nil {
		return ChatDetail{}, err
	}
	// Siswa membuka chat → tandai semua balasan BK terbaca (reset badge).
	_ = c.queries.ResetStudentUnread(ctx, chat.ID)
	msgs, err := c.queries.ListMessagesByChat(ctx, db.ListMessagesByChatParams{ChatID: chat.ID, Limit: 200, Offset: 0})
	if err != nil {
		return ChatDetail{}, utils.NewErrInternal("gagal memuat pesan")
	}
	views, err := fillAttachments(ctx, c.queries, c.messagesToView(msgs), msgs)
	if err != nil {
		return ChatDetail{}, err
	}
	return ChatDetail{
		Chat:     ChatView{ID: chat.ID.String(), IdentityMode: chat.IdentityMode, Status: chat.Status},
		Messages: views,
	}, nil
}

func (c *Chat) messagesToView(msgs []db.Message) []MessageView {
	out := make([]MessageView, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, MessageView{
			ID: m.ID.String(), Sender: m.Sender, Body: m.Body,
			Attachments: []AttachmentView{}, CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func (c *Chat) Reply(ctx context.Context, session db.AnonSession, chatID uuid.UUID, in ReplyInput) (MessageView, error) {
	chat, err := c.ownedChat(ctx, session, chatID)
	if err != nil {
		return MessageView{}, err
	}
	if chat.HiddenByStaffAt != nil {
		return MessageView{}, utils.NewErrChatClosedByStaff("Chat ditutup oleh Guru BK.")
	}
	body := strings.TrimSpace(in.Body)
	if len([]rune(body)) > 2000 {
		return MessageView{}, utils.NewErrInvalidInput("Pesan maksimal 2.000 karakter.")
	}
	if body == "" && len(in.Files) == 0 {
		return MessageView{}, utils.NewErrInvalidInput("Tulis pesan sebelum mengirim.")
	}
	if err := c.rl.CheckMessage(session.ID.String(), in.ClientIP); err != nil {
		return MessageView{}, err
	}
	saved, err := c.saveFiles(in.Files)
	if err != nil {
		return MessageView{}, err
	}
	// F8: pesan + lampiran + requeue (status/touch) dalam SATU transaksi.
	var view MessageView
	err = db.WithTx(ctx, c.pool, func(q *db.Queries) error {
		msg, err := q.CreateMessage(ctx, db.CreateMessageParams{
			ChatID: chat.ID, Sender: "SISWA", SenderUserID: nil, Body: body,
		})
		if err != nil {
			return utils.NewErrInternal("gagal menyimpan pesan")
		}
		if err := attachFilesTx(ctx, q, msg.ID, saved); err != nil {
			return err
		}
		// requeue: status balik 'baru' (keputusan #4/#55) + touch last_message_at
		if _, err := q.UpdateChatStatus(ctx, db.UpdateChatStatusParams{ID: chat.ID, Status: "baru"}); err != nil {
			return utils.NewErrInternal("gagal memperbarui status")
		}
		if _, err := q.TouchChatMessage(ctx, chat.ID); err != nil {
			return utils.NewErrInternal("gagal memperbarui waktu")
		}
		view = MessageView{ID: msg.ID.String(), Sender: "SISWA", Body: body, Attachments: []AttachmentView{}, CreatedAt: msg.CreatedAt.Format(time.RFC3339)}
		return nil
	})
	if err != nil {
		c.discardFiles(saved)
		return MessageView{}, err
	}
	return view, nil
}

func (c *Chat) HideByStudent(ctx context.Context, session db.AnonSession, chatID uuid.UUID) error {
	if _, err := c.ownedChat(ctx, session, chatID); err != nil {
		return err
	}
	if _, err := c.queries.HideByStudent(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal menghapus laporan")
	}
	return nil
}

// AttachmentMeta dikembalikan untuk serve file.
type AttachmentMeta struct {
	StoredPath   string
	OriginalName string
	Mime         string
	ChatID       uuid.UUID
}

func (c *Chat) AttachmentMeta(ctx context.Context, id uuid.UUID) (AttachmentMeta, error) {
	a, err := c.queries.GetAttachmentByID(ctx, id)
	if err != nil {
		return AttachmentMeta{}, utils.NewErrNotFound("Lampiran tidak ditemukan.")
	}
	msg, err := c.queries.GetMessageByID(ctx, a.MessageID)
	if err != nil {
		return AttachmentMeta{}, utils.NewErrNotFound("Lampiran tidak ditemukan.")
	}
	name := "lampiran"
	if a.OriginalName != nil {
		name = *a.OriginalName
	}
	return AttachmentMeta{StoredPath: a.StoredPath, OriginalName: name, Mime: a.Mime, ChatID: msg.ChatID}, nil
}

// Store mengembalikan store untuk membaca file.
func (c *Chat) Store() *storage.Store { return c.store }
