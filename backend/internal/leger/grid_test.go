package leger

import (
	"strings"
	"testing"
)

func TestGridFromCSV(t *testing.T) {
	g, err := GridFromCSV(strings.NewReader("a,b\nc,d\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(g.Rows) != 2 || g.Rows[0][0] != "a" || g.Rows[1][1] != "d" {
		t.Fatalf("bad grid: %#v", g.Rows)
	}
}

func TestGridFromCSVSkipsBlankLines(t *testing.T) {
	g, err := GridFromCSV(strings.NewReader("a,b\n\nc,d\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(g.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(g.Rows))
	}
}
