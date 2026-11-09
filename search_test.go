package pagination

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchTermReadsFirstNonEmptyKey(t *testing.T) {
	r := httptest.NewRequest("GET", "/items?search=%20%20&q=%20pizza%20%20margherita%20", nil)
	if got := SearchTerm(r); got != "pizza margherita" {
		t.Fatalf("got %q", got)
	}
	r = httptest.NewRequest("GET", "/items?term=abc", nil)
	if got := SearchTerm(r); got != "abc" {
		t.Fatalf("got %q", got)
	}
	r = httptest.NewRequest("GET", "/items?name=abc", nil)
	if got := SearchTerm(r, "name"); got != "abc" {
		t.Fatalf("custom key: got %q", got)
	}
	if got := SearchTerm(httptest.NewRequest("GET", "/items", nil)); got != "" {
		t.Fatalf("empty: got %q", got)
	}
}

func TestNormalizeSearchCapsRunes(t *testing.T) {
	long := strings.Repeat("é", 150)
	got := NormalizeSearch(long)
	if n := len([]rune(got)); n != MaxSearchLength {
		t.Fatalf("want %d runes, got %d", MaxSearchLength, n)
	}
}

func TestMeetsMinLength(t *testing.T) {
	if MeetsMinLength("ab", 3) || !MeetsMinLength("abc", 3) || !MeetsMinLength("ñoñ", 3) {
		t.Fatal("rune-based min length is wrong")
	}
}

func TestEscapeLikeMatchesLiterally(t *testing.T) {
	cases := map[string]string{
		"50%":     `50\%`,
		"a_b":     `a\_b`,
		`c:\temp`: `c:\\temp`,
		"plain":   "plain",
		`%_\`:     `\%\_\\`,
	}
	for in, want := range cases {
		if got := EscapeLike(in); got != want {
			t.Errorf("EscapeLike(%q) = %q, want %q", in, got, want)
		}
	}
	if got := ContainsPattern("10%"); got != `%10\%%` {
		t.Fatalf("ContainsPattern = %q", got)
	}
}
