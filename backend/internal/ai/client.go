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
2. Jika suatu data tidak tersedia, jangan disimpulkan.
3. JANGAN memberi diagnosis medis/psikologis, skor risiko, atau prediksi
   (kelulusan, perilaku, dll).
4. Gunakan bahasa Indonesia yang jelas dan tidak menghakimi.

Balas HANYA dengan satu objek JSON valid (tanpa markdown/penjelasan di luar JSON)
dengan skema PERSIS berikut — HANYA dua field:
{
  "headline": "satu kalimat inti tentang siswa",
  "narrative": "paragraf naratif 2-4 kalimat yang merangkum temuan utama"
}`
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

// UserPrompt minta model membalas JSON; dikirim sebagai pesan user.
func BuildUserPromptJSON(in SummarizeInput) string {
	return BuildUserPrompt(in) + `

PENTING: balas HANYA objek JSON valid (mulai dengan "{" dan diakhiri "}"),
tanpa penjelasan, tanpa blok markdown, tanpa teks di luar JSON.`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	Temperature    float64       `json:"temperature"`
	MaxTokens      int           `json:"max_tokens"`
	ResponseFormat *respFormat   `json:"response_format,omitempty"`
}

// respFormat: minta provider OpenAI-compatible membalas JSON valid.
type respFormat struct {
	Type string `json:"type"` // "json_object"
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Narrative: hasil ringkas dari AI (headline + narrative).
type Narrative struct {
	Headline  string `json:"headline"`
	Narrative string `json:"narrative"`
}

// Summarize memanggil {AI_BASE_URL}/chat/completions (OpenAI-compatible) dan
// mengembalikan headline+narrative. Bila model membalas non-JSON, coba sekali
// lagi dengan penegasan lebih keras; bila tetap gagal, pakai teks mentah
// sebagai narrative (headline kosong).
func (c *Client) Summarize(ctx context.Context, in SummarizeInput) (Narrative, error) {
	raw, err := c.complete(ctx, BuildUserPrompt(in))
	if err != nil {
		return Narrative{}, err
	}
	if n, ok := parseNarrative(raw); ok {
		return n, nil
	}
	raw2, err2 := c.complete(ctx, BuildUserPromptJSON(in))
	if err2 != nil {
		// Pakai hasil pertama apa adanya sebagai narrative.
		return Narrative{Narrative: strings.TrimSpace(raw)}, nil
	}
	if n, ok := parseNarrative(raw2); ok {
		return n, nil
	}
	return Narrative{Narrative: strings.TrimSpace(raw2)}, nil
}

// parseNarrative: ekstrak headline+narrative dari balasan model. Model sering
// TIDAK mengikuti skema (mis. membalas {"ringkasan_profil": {...}} atau
// markdown). Strategi berlapis:
//  1. JSON persis {headline, narrative}.
//  2. JSON apa pun: cari key "headline"/"narrative"/"kesimpulan" secara rekursif.
//  3. Teks bebas: pakai seluruh teks sebagai narrative (headline kosong).
func parseNarrative(s string) (Narrative, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Narrative{}, false
	}
	// 1. JSON persis.
	var direct Narrative
	if json.Unmarshal([]byte(s), &direct) == nil {
		if strings.TrimSpace(direct.Narrative) != "" || strings.TrimSpace(direct.Headline) != "" {
			return direct, true
		}
	}
	// 2. JSON apa pun → cari key relevan secara rekursif.
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			var anyJSON any
			if json.Unmarshal([]byte(s[i:j+1]), &anyJSON) == nil {
				n := Narrative{
					Headline:  findStringKey(anyJSON, "headline", "judul"),
					Narrative: findStringKey(anyJSON, "narrative", "kesimpulan", "ringkasan", "analisis"),
				}
				if strings.TrimSpace(n.Narrative) != "" || strings.TrimSpace(n.Headline) != "" {
					return n, true
				}
			}
		}
	}
	// 3. Teks bebas → narrative (bersihkan markdown agar rapi).
	return Narrative{Narrative: stripMarkdown(s)}, true
}

// stripMarkdown: bersihkan penanda markdown umum (heading, bold, tabel, list)
// supaya narrative terbaca sebagai teks biasa di UI. Baris tabel dibuang karena
// angka sudah tersedia di metrics/data_used.
// isHorizontalRule: baris `---` / `***` / `___` (garis horizontal markdown)
// yang sering muncul di jawaban model — dibuang agar tidak tampil di UI.
func isHorizontalRule(t string) bool {
	if len(t) < 3 {
		return false
	}
	c := t[0]
	if c != '-' && c != '*' && c != '_' {
		return false
	}
	for i := 0; i < len(t); i++ {
		if t[i] != c {
			return false
		}
	}
	return true
}

func stripMarkdown(s string) string {
	var lines []string
	for _, ln := range strings.Split(s, "\n") {
		t := strings.TrimSpace(ln)
		if strings.Contains(t, "|") { // baris tabel markdown → buang
			continue
		}
		if isHorizontalRule(t) { // garis horizontal --- / *** / ___
			continue
		}
		t = strings.TrimLeft(t, "#") // heading
		t = strings.TrimSpace(t)
		t = strings.ReplaceAll(t, "**", "")
		t = strings.ReplaceAll(t, "__", "")
		t = strings.TrimPrefix(t, "- ") // bullet
		t = strings.TrimPrefix(t, "* ")
		lines = append(lines, t)
	}
	out := strings.Join(lines, "\n")
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(out)
}

// findStringKey: cari nilai string pertama untuk salah satu key (rekursif).
func findStringKey(v any, keys ...string) string {
	keySet := map[string]bool{}
	for _, k := range keys {
		keySet[k] = true
	}
	var walk func(any) string
	walk = func(x any) string {
		switch t := x.(type) {
		case map[string]any:
			for k, val := range t {
				if keySet[k] {
					if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
						return s
					}
				}
			}
			for _, val := range t {
				if r := walk(val); r != "" {
					return r
				}
			}
		case []any:
			for _, val := range t {
				if r := walk(val); r != "" {
					return r
				}
			}
		}
		return ""
	}
	return walk(v)
}

func (c *Client) complete(ctx context.Context, userPrompt string) (string, error) {
	body, _ := json.Marshal(chatRequest{
		Model:          c.model,
		Messages:       []chatMessage{{Role: "system", Content: SystemPrompt()}, {Role: "user", Content: userPrompt}},
		Temperature:    0.3,
		MaxTokens:      2000,
		ResponseFormat: &respFormat{Type: "json_object"},
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
	// Toleran: sebagian proxy mengirim JSON + konten tambahan setelahnya
	// (mis. chunk usage). Decoder berhenti setelah nilai JSON pertama.
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&out); err != nil {
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
