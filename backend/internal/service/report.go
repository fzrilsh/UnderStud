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
	Mode  string `json:"mode"`
	Name  string `json:"name,omitempty"`
	Kelas string `json:"kelas,omitempty"`
}

type ReportItem struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	Pinned        bool           `json:"pinned"`
	Drafted       bool           `json:"drafted"`
	Identity      ReportIdentity `json:"identity"`
	LastMessageAt string         `json:"last_message_at"`
}

type ReportDetail struct {
	Chat     ReportItem    `json:"chat"`
	Messages []MessageView `json:"messages"`
}

func toReportItem(ch db.Chat) ReportItem {
	ident := ReportIdentity{Mode: ch.IdentityMode}
	if ch.IdentityMode == "profil" {
		if ch.SnapshotName != nil {
			ident.Name = *ch.SnapshotName
		}
		if ch.SnapshotKelas != nil {
			ident.Kelas = *ch.SnapshotKelas
		}
	}
	return ReportItem{
		ID: ch.ID.String(), Status: ch.Status,
		Pinned: ch.PinnedAt != nil, Drafted: ch.DraftedAt != nil,
		Identity: ident, LastMessageAt: ch.LastMessageAt.Format(time.RFC3339),
	}
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
	out := make([]ReportItem, 0, len(chats))
	for _, ch := range chats {
		out = append(out, toReportItem(ch))
	}
	return out, total, nil
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
	out := make([]ReportItem, 0, len(chats))
	for _, ch := range chats {
		out = append(out, toReportItem(ch))
	}
	return out, total, nil
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
	return ReportDetail{Chat: toReportItem(ch), Messages: views}, nil
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
	if _, err := r.get(ctx, chatID); err != nil {
		return err
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

// Next = keluar dari antrean: lepas draft + status 'dilewati' (fix audit A1).
// Syarat: BK sudah membalas ≥1× (keputusan #40/#53).
func (r *Report) Next(ctx context.Context, chatID uuid.UUID) error {
	if _, err := r.get(ctx, chatID); err != nil {
		return err
	}
	if err := r.requireReplied(ctx, chatID); err != nil {
		return err
	}
	if _, err := r.queries.UndraftChat(ctx, chatID); err != nil {
		return utils.NewErrInternal("gagal melewati laporan")
	}
	if _, err := r.queries.UpdateChatStatus(ctx, db.UpdateChatStatusParams{ID: chatID, Status: "dilewati"}); err != nil {
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
	return MessageView{ID: msg.ID.String(), Sender: "BK", Body: body, Attachments: []AttachmentView{}, CreatedAt: msg.CreatedAt.Format(time.RFC3339)}, nil
}
