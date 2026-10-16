package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/db"
	"backend/internal/utils"
)

type Report struct {
	queries *db.Queries
}

func NewReport(q *db.Queries) *Report { return &Report{queries: q} }

type ReportIdentity struct {
	Mode  string  `json:"mode"`
	Name  *string `json:"name"`  // M11: selalu ada (null bila anonim)
	Kelas *string `json:"kelas"` // M11
}

type ReportItem struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	Pinned        bool           `json:"pinned"`
	Drafted       bool           `json:"drafted"`
	PinnedAt      *string        `json:"pinned_at"`  // M11: timestamp sesuai spec §21.4
	DraftedAt     *string        `json:"drafted_at"` // M11
	Identity      ReportIdentity `json:"identity"`
	LastMessageAt string         `json:"last_message_at"`
	LastMessage   *LastMessage   `json:"last_message"`    // preview pesan terakhir
	UnreadByStaff bool           `json:"unread_by_staff"` // F23: pesan terakhir dari siswa
}

// LastMessage: preview pesan terakhir untuk daftar chat.
type LastMessage struct {
	Sender    string `json:"sender"` // SISWA | BK
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type ReportDetail struct {
	Chat     ReportItem    `json:"chat"`
	Messages []MessageView `json:"messages"`
}

func toReportItem(ch db.Chat, unread bool) ReportItem {
	ident := ReportIdentity{Mode: ch.IdentityMode}
	if ch.IdentityMode == "profil" {
		ident.Name = ch.SnapshotName
		ident.Kelas = ch.SnapshotKelas
	}
	item := ReportItem{
		ID: ch.ID.String(), Status: ch.Status,
		Pinned: ch.PinnedAt != nil, Drafted: ch.DraftedAt != nil,
		Identity: ident, LastMessageAt: ch.LastMessageAt.Format(time.RFC3339),
		UnreadByStaff: unread,
	}
	if ch.PinnedAt != nil {
		s := ch.PinnedAt.Format(time.RFC3339)
		item.PinnedAt = &s
	}
	if ch.DraftedAt != nil {
		s := ch.DraftedAt.Format(time.RFC3339)
		item.DraftedAt = &s
	}
	return item
}

func (r *Report) Queue(ctx context.Context, limit, offset int32) ([]ReportItem, int64, error) {
	chats, err := r.queries.ListStaffQueue(ctx, db.ListStaffQueueParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat antrean")
	}
	total, err := r.queries.CountStaffQueue(ctx)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat antrean")
	}
	return r.withUnread(ctx, chats), total, nil
}

// History = riwayat: laporan 'dilewati' atau di-hide BK (spec §7.2, §13).
func (r *Report) History(ctx context.Context, limit, offset int32) ([]ReportItem, int64, error) {
	chats, err := r.queries.ListStaffHistory(ctx, db.ListStaffHistoryParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat riwayat")
	}
	total, err := r.queries.CountStaffHistory(ctx)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat riwayat")
	}
	return r.withUnread(ctx, chats), total, nil
}

// withUnread: bikin ReportItem + unread_by_staff (F23) + preview pesan terakhir
// lewat satu query batch pengirim/pesan terakhir per chat.
func (r *Report) withUnread(ctx context.Context, chats []db.Chat) []ReportItem {
	out := make([]ReportItem, 0, len(chats))
	if len(chats) == 0 {
		return out
	}
	ids := make([]uuid.UUID, 0, len(chats))
	for _, ch := range chats {
		ids = append(ids, ch.ID)
	}
	lastByChat := map[uuid.UUID]LastMessage{}
	if rows, err := r.queries.ListLastMessages(ctx, ids); err == nil {
		for _, row := range rows {
			lastByChat[row.ChatID] = LastMessage{
				Sender: row.Sender, Body: row.Body, CreatedAt: row.CreatedAt.Format(time.RFC3339),
			}
		}
	}
	for _, ch := range chats {
		item := toReportItem(ch, lastByChat[ch.ID].Sender == "SISWA")
		if lm, ok := lastByChat[ch.ID]; ok {
			item.LastMessage = &lm
		}
		out = append(out, item)
	}
	return out
}

func (r *Report) get(ctx context.Context, chatID uuid.UUID) (db.Chat, error) {
	ch, err := r.queries.GetChatByID(ctx, chatID)
	if err != nil {
		return db.Chat{}, utils.NewErrNotFound("Laporan tidak ditemukan.")
	}
	return ch, nil
}

func (r *Report) Detail(ctx context.Context, chatID uuid.UUID) (ReportDetail, error) {
	ch, err := r.get(ctx, chatID)
	if err != nil {
		return ReportDetail{}, err
	}
	// H2: BK membuka laporan → transisi status baru→dibuka (spec §7.1).
	if opened, err := r.queries.MarkChatOpened(ctx, chatID); err == nil {
		ch = opened
	}
	msgs, err := r.queries.ListMessagesByChat(ctx, db.ListMessagesByChatParams{ChatID: ch.ID, Limit: 200, Offset: 0})
	if err != nil {
		return ReportDetail{}, utils.NewErrInternal("gagal memuat pesan")
	}
	views := make([]MessageView, 0, len(msgs))
	for _, m := range msgs {
		views = append(views, MessageView{
			ID: m.ID.String(), Sender: m.Sender, Body: m.Body,
			Attachments: []AttachmentView{}, CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}
	views, err = fillAttachments(ctx, r.queries, views, msgs)
	if err != nil {
		return ReportDetail{}, err
	}
	unread := false
	item := toReportItem(ch, unread)
	if len(msgs) > 0 {
		last := msgs[len(msgs)-1]
		unread = last.Sender == "SISWA"
		item.UnreadByStaff = unread
		item.LastMessage = &LastMessage{
			Sender: last.Sender, Body: last.Body, CreatedAt: last.CreatedAt.Format(time.RFC3339),
		}
	}
	return ReportDetail{Chat: item, Messages: views}, nil
}

// requireReplied menegakkan syarat "balas ≥1×" (keputusan #40/#53).
func (r *Report) requireReplied(ctx context.Context, chatID uuid.UUID) error {
	ok, err := r.queries.HasStaffReplied(ctx, chatID)
	if err != nil {
		return utils.NewErrInternal("gagal memeriksa balasan")
	}
	if !ok {
		return utils.NewErrInvalidInput("Balas dulu sebelum melewati.")
	}
	return nil
}

func (r *Report) Pin(ctx context.Context, chatID uuid.UUID) error {
	ch, err := r.get(ctx, chatID)
	if err != nil {
		return err
	}
	// M7: Pin dari Draft wajib balas ≥1× (spec §7.2/#53). Dari Antrean bebas.
	if ch.DraftedAt != nil {
		if err := r.requireReplied(ctx, chatID); err != nil {
			return err
		}
	}
	if _, err := r.queries.PinChat(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal menyematkan")
	}
	return nil
}

func (r *Report) Unpin(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.UnpinChat(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal melepas sematan")
	}
	return nil
}

func (r *Report) Draft(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.DraftChat(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal menyimpan draft")
	}
	return nil
}

// Undraft = keluarkan dari draft (tombol "Keluarkan dari draft"). Status jadi
// 'dilewati' agar keluar dari section Draft; pin tetap dipertahankan bila ada.
func (r *Report) Undraft(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.SetChatSkipped(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal mengeluarkan dari draft")
	}
	return nil
}

// Skip = lewati chat ke antrean berikutnya TANPA wajib balas (tombol
// "Berikutnya" saat belum dibalas). Status 'dilewati' + lepas draft;
// sematan dipertahankan (lihat SetChatSkipped).
func (r *Report) Skip(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.SetChatSkipped(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal melewati laporan")
	}
	return nil
}

// Next = "Selesai & lanjut": keluar dari antrean (status 'dilewati') + lepas
// draft, tapi **pin dipertahankan** agar chat yang disematkan tetap ada di
// section "Disematkan". Syarat: BK sudah membalas ≥1× (keputusan #40/#53).
func (r *Report) Next(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if err := r.requireReplied(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.SetChatSkipped(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal melewati laporan")
	}
	return nil
}

func (r *Report) Hide(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.HideByStaff(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal menghapus laporan")
	}
	return nil
}

func (r *Report) Reply(ctx context.Context, staffID, chatID uuid.UUID, body string) (MessageView, error) {
	ch, err := r.get(ctx, chatID)
	if err != nil {
		return MessageView{}, err
	}
	if ch.HiddenByStudentAt != nil {
		return MessageView{}, utils.NewErrChatHiddenByStudent("Chat ini disembunyikan oleh siswa.")
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return MessageView{}, utils.NewErrInvalidInput("Tulis balasan sebelum mengirim.")
	}
	if len([]rune(body)) > 2000 {
		return MessageView{}, utils.NewErrInvalidInput("Balasan maksimal 2.000 karakter.")
	}
	msg, err := r.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChatID: ch.ID, Sender: "BK", SenderUserID: &staffID, Body: body,
	})
	if err != nil {
		return MessageView{}, utils.NewErrInternal("gagal menyimpan balasan")
	}
	if _, err := r.queries.TouchChatMessage(ctx, ch.ID); err != nil {
		return MessageView{}, utils.NewErrInternal("gagal memperbarui waktu")
	}
	// Tandai belum dibaca sisi siswa (badge di sidebar siswa).
	if err := r.queries.IncrementStudentUnread(ctx, ch.ID); err != nil {
		return MessageView{}, utils.NewErrInternal("gagal memperbarui status baca")
	}
	return MessageView{ID: msg.ID.String(), Sender: "BK", Body: body, Attachments: []AttachmentView{}, CreatedAt: msg.CreatedAt.Format(time.RFC3339)}, nil
}
