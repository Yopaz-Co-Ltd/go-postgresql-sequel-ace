package main

import "testing"

func TestParseLimit(t *testing.T) {
	tests := map[string]int{
		"":     200,
		"abc":  200,
		"0":    200,
		"25":   25,
		"5000": 1000,
	}

	for input, want := range tests {
		if got := parseLimit(input); got != want {
			t.Fatalf("parseLimit(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestQuoteIdent(t *testing.T) {
	got := quoteIdent(`public"user`)
	want := `"public""user"`
	if got != want {
		t.Fatalf("quoteIdent mismatch: got %q want %q", got, want)
	}
}
