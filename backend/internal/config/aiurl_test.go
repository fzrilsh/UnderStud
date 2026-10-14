package config

import "testing"

func TestNormalizeAIBaseURL(t *testing.T) {
	// Tes ini berjalan DI HOST (go test), jadi inContainer=false.
	// host.docker.internal → 127.0.0.1 (tak resolve di host).
	cases := map[string]string{
		"http://host.docker.internal:20128/v1": "http://127.0.0.1:20128/v1",
		"http://localhost:11434/v1":            "http://localhost:11434/v1", // di host, dibiarkan
		"http://127.0.0.1:20128/v1":            "http://127.0.0.1:20128/v1",
		"https://api.openai.com/v1":            "https://api.openai.com/v1", // host lain: apa adanya
		"":                                     "",
	}
	for in, want := range cases {
		if got := normalizeAIBaseURL(in); got != want {
			t.Errorf("normalizeAIBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}
