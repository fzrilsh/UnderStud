package service

import (
	"strings"
	"testing"
	"time"
)

func TestRenderTXTContainsHeaderAndBody(t *testing.T) {
	out := string(renderTXT("Andi Pratama", "XI-IPA-1", "Kemampuan: baik.", time.Unix(0, 0)))
	for _, want := range []string{"Andi Pratama", "XI-IPA-1", "Kemampuan: baik.", "bukan diagnosis"} {
		if !strings.Contains(out, want) {
			t.Fatalf("TXT missing %q:\n%s", want, out)
		}
	}
}

func TestRenderPDFIsPDF(t *testing.T) {
	out, err := renderPDF("Andi", "XI-IPA-1", "Isi ringkasan.", time.Unix(0, 0))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(out) < 5 || string(out[:5]) != "%PDF-" {
		t.Fatalf("not a PDF (len=%d)", len(out))
	}
}

func TestRenderPDFEmptyContent(t *testing.T) {
	if _, err := renderPDF("Andi", "XI-IPA-1", "", time.Unix(0, 0)); err != nil {
		t.Fatalf("empty content should still render: %v", err)
	}
}
