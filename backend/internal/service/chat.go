package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

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
	store   *storage.Store
	rl      *middleware.RateLimiter
	cfg     config.Config
}

func NewChat(q *db.Queries, store *storage.Store, rl *middleware.RateLimiter, cfg config.Config) *Chat {
	return &Chat{queries: q, store: store, rl: rl, cfg: cfg}
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
}

type ReplyInput struct {
	Body  string
	Files []UploadFile
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
	ID            string `json:"id"`
	IdentityMode  string `json:"identity_mode"`
	Status        string `json:"status"`
	LastMessageAt string `json:"last_message_at"`
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
	if lat == 0 && lng == 0 {
		return utils.NewErrLocationUnavailable("Aktifkan izin lokasi untuk mengirim laporan.")
	}
	dist := geo.DistanceM(lat, lng, c.cfg.SchoolLat, c.cfg.SchoolLng)
	if dist > float64(c.cfg.SchoolRadiusM) {
		return utils.NewErrOutOfRange("Kamu di luar area sekolah. Laporan hanya bisa dikirim dari lingkungan sekolah.")
	}
	if accuracy > float64(c.cfg.SchoolRadiusM) {
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
	if err := c.rl.CheckChat(session.ID.String()); err != nil {
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

	chat, err := c.queries.CreateChat(ctx, db.CreateChatParams{
		SessionID:     session.ID,
		IdentityMode:  mode,
		SnapshotName:  snapName,
		SnapshotKelas: snapKelas,
	})
	if err != nil {
		return ChatView{}, utils.NewErrInternal("gagal membuat laporan")
	}
	msg, err := c.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChatID: chat.ID, Sender: "SISWA", SenderUserID: nil, Body: body,
	})
	if err != nil {
		return ChatView{}, utils.NewErrInternal("gagal menyimpan pesan")
	}
	if err := c.attachFiles(ctx, msg.ID, saved); err != nil {
		return ChatView{}, err
	}
	return ChatView{ID: chat.ID.String(), IdentityMode: chat.IdentityMode, Status: chat.Status}, nil
}

// saveFiles memvalidasi & menyimpan; mengembalikan daftar file tersimpan.
func (c *Chat) saveFiles(files []UploadFile) ([]storage.SavedFile, error) {
	if len(files) > 3 {
		return nil, utils.NewErrInvalidInput("Maksimal 3 file per pesan.")
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
			return nil, utils.NewErrInternal("gagal menyimpan file")
		}
		total += sf.Size
		if total > c.cfg.MaxUploadTotalMB*1024*1024 {
			return nil, utils.NewErrPayloadTooLarge("Total lampiran maksimal 20 MB per pesan.")
		}
		out = append(out, sf)
	}
	return out, nil
}

func (c *Chat) attachFiles(ctx context.Context, messageID uuid.UUID, saved []storage.SavedFile) error {
	for _, sf := range saved {
		name := sf.StoredPath
		if _, err := c.queries.CreateAttachment(ctx, db.CreateAttachmentParams{
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

func (c *Chat) ListBySession(ctx context.Context, session db.AnonSession, limit, offset int32) ([]ChatSummary, int64, error) {
	chats, err := c.queries.ListChatsBySession(ctx, db.ListChatsBySessionParams{
		SessionID: session.ID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat daftar")
	}
	out := make([]ChatSummary, 0, len(chats))
	for _, ch := range chats {
		out = append(out, ChatSummary{
			ID: ch.ID.String(), IdentityMode: ch.IdentityMode, Status: ch.Status,
			LastMessageAt: ch.LastMessageAt.Format(time.RFC3339),
		})
	}
	return out, int64(len(out)), nil
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
	msgs, err := c.queries.ListMessagesByChat(ctx, db.ListMessagesByChatParams{ChatID: chat.ID, Limit: 200, Offset: 0})
	if err != nil {
		return ChatDetail{}, utils.NewErrInternal("gagal memuat pesan")
	}
	return ChatDetail{
		Chat:     ChatView{ID: chat.ID.String(), IdentityMode: chat.IdentityMode, Status: chat.Status},
		Messages: c.messagesToView(msgs),
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
	if err := c.rl.CheckMessage(session.ID.String()); err != nil {
		return MessageView{}, err
	}
	saved, err := c.saveFiles(in.Files)
	if err != nil {
		return MessageView{}, err
	}
	msg, err := c.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChatID: chat.ID, Sender: "SISWA", SenderUserID: nil, Body: body,
	})
	if err != nil {
		return MessageView{}, utils.NewErrInternal("gagal menyimpan pesan")
	}
	if err := c.attachFiles(ctx, msg.ID, saved); err != nil {
		return MessageView{}, err
	}
	// requeue: status balik 'baru' (keputusan #4/#55) + touch last_message_at
	if _, err := c.queries.UpdateChatStatus(ctx, db.UpdateChatStatusParams{ID: chat.ID, Status: "baru"}); err != nil {
		return MessageView{}, utils.NewErrInternal("gagal memperbarui status")
	}
	if _, err := c.queries.TouchChatMessage(ctx, chat.ID); err != nil {
		return MessageView{}, utils.NewErrInternal("gagal memperbarui waktu")
	}
	return MessageView{ID: msg.ID.String(), Sender: "SISWA", Body: body, Attachments: []AttachmentView{}, CreatedAt: msg.CreatedAt.Format(time.RFC3339)}, nil
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
