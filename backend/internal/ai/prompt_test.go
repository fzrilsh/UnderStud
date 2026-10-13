package ai

import (
	"strings"
	"testing"
)

func TestBuildUserPromptIncludesFilled(t *testing.T) {
	p := BuildUserPrompt(SummarizeInput{Kelas: "XI-IPA-1", MinatBakat: "basket"})
	if !strings.Contains(p, "XI-IPA-1") || !strings.Contains(p, "basket") {
		t.Fatalf("prompt missing fields: %s", p)
	}
}

func TestBuildUserPromptMarksEmpty(t *testing.T) {
	p := BuildUserPrompt(SummarizeInput{Kelas: "XI-IPA-1"})
	if !strings.Contains(p, "tidak tersedia") {
		t.Fatalf("empty fields should be marked: %s", p)
	}
}

func TestSystemPromptForbidsDiagnosis(t *testing.T) {
	s := SystemPrompt()
	if !strings.Contains(s, "BUKAN mendiagnosis") {
		t.Fatal("system prompt must forbid diagnosis")
	}
}
