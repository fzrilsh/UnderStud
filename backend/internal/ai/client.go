package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
)

type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.AIBaseURL, "/"),
		apiKey:  cfg.AIAPIKey,
		model:   cfg.AIModel,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// Configured: fitur aktif hanya bila AI_API_KEY terisi (spec §9).
func (c *Client) Configured() bool { return strings.TrimSpace(c.apiKey) != "" }

type SummarizeInput struct {
	Kelas                string
	Kehadiran            string
	NilaiErapor          string
	MinatBakat           string
	CatatanPsikologis    string
	CatatanKasusPrestasi string
	MasukanBK            string
	PendapatWaliKelas    string
	Observasi            []string
}

func SystemPrompt() string {
	return `Kamu asisten yang membantu Guru BK dan Wali Kelas memahami kondisi siswa.
Tugasmu meringkas data yang diberikan — BUKAN mendiagnosis.

Aturan wajib:
1. Ringkas HANYA berdasarkan data yang diberikan. Jangan mengarang.
2. Jika suatu data tidak tersedia, tulis "tidak tersedia" — jangan disimpulkan.
3. Sertakan sumber untuk setiap klaim (mis. "berdasarkan nilai e-rapor").
4. JANGAN memberi diagnosis medis/psikologis, skor risiko, atau prediksi
   (kelulusan, perilaku, dll).
5. Gunakan bahasa Indonesia yang jelas dan tidak menghakimi.
6. Susun dalam 4 bagian: Kemampuan, Minat, Kekuatan, Kebutuhan.`
}

func orNA(s string) string {
	if strings.TrimSpace(s) == "" {
		return "tidak tersedia"
	}
	return s
}

func BuildUserPrompt(in SummarizeInput) string {
	obs := "tidak tersedia"
	if len(in.Observasi) > 0 {
		obs = strings.Join(in.Observasi, "\n")
	}
	return fmt.Sprintf(`Data siswa:
- Kelas: %s
- Kehadiran: %s
- Nilai e-rapor: %s
- Minat & bakat: %s
- Catatan psikologis: %s
- Catatan kasus/prestasi: %s
- Masukan BK: %s
- Pendapat Wali Kelas: %s
- Observasi terbaru (%d catatan):
%s

Buat ringkasan profil siswa sesuai aturan.`,
		orNA(in.Kelas), orNA(in.Kehadiran), orNA(in.NilaiErapor), orNA(in.MinatBakat),
		orNA(in.CatatanPsikologis), orNA(in.CatatanKasusPrestasi),
		orNA(in.MasukanBK), orNA(in.PendapatWaliKelas),
		len(in.Observasi), obs)
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Summarize memanggil {AI_BASE_URL}/chat/completions (OpenAI-compatible).
func (c *Client) Summarize(ctx context.Context, in SummarizeInput) (string, error) {
	body, _ := json.Marshal(chatRequest{
		Model:       c.model,
		Messages:    []chatMessage{{Role: "system", Content: SystemPrompt()}, {Role: "user", Content: BuildUserPrompt(in)}},
		Temperature: 0.3,
		MaxTokens:   800,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ai http %d: %s", resp.StatusCode, string(raw))
	}
	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("ai error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("ai returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
