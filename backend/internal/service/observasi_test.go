package service

import "testing"

func TestValidateItemsValid(t *testing.T) {
	items, cats, red, err := ValidateItems([]string{"emosi:murung", "sosial:menyendiri"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if len(cats) != 2 || cats[0] != "emosi" || cats[1] != "sosial" {
		t.Fatalf("bad categories: %v", cats)
	}
	if red {
		t.Fatal("should not be red flag")
	}
}

func TestValidateItemsRedFlag(t *testing.T) {
	_, _, red, err := ValidateItems([]string{"red_flag:menyakiti_diri"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !red {
		t.Fatal("want red flag true")
	}
}

func TestValidateItemsTidakAdaExclusive(t *testing.T) {
	_, _, _, err := ValidateItems([]string{"tidak_ada", "emosi:murung"})
	if err == nil {
		t.Fatal("want error: tidak_ada eksklusif")
	}
}

func TestValidateItemsUnknown(t *testing.T) {
	_, _, _, err := ValidateItems([]string{"emosi:tidak_ada_di_catalog"})
	if err == nil {
		t.Fatal("want error for unknown item")
	}
}

func TestValidateItemsEmpty(t *testing.T) {
	_, _, _, err := ValidateItems(nil)
	if err == nil {
		t.Fatal("want error for empty items")
	}
}
