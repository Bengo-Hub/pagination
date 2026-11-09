package pagination

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

// MaxSearchLength caps a search term in runes. The UI applies the same cap
// (shared-ui-lib normalizeSearchTerm), so a longer term only arrives from scripts.
const MaxSearchLength = 100

// defaultSearchKeys are the query keys read by SearchTerm when none are given,
// in priority order. Services use all three names today.
var defaultSearchKeys = []string{"search", "q", "term"}

// SearchTerm returns the normalised search term from the first non-empty query
// key (default: search, q, term). It trims, collapses runs of whitespace and caps
// the length, matching the UI rule, so an untrimmed or padded term never reaches SQL.
func SearchTerm(r *http.Request, keys ...string) string {
	if len(keys) == 0 {
		keys = defaultSearchKeys
	}
	q := r.URL.Query()
	for _, k := range keys {
		if v := NormalizeSearch(q.Get(k)); v != "" {
			return v
		}
	}
	return ""
}

// NormalizeSearch trims, collapses whitespace and caps the term at MaxSearchLength runes.
func NormalizeSearch(raw string) string {
	term := strings.Join(strings.Fields(raw), " ")
	if utf8.RuneCountInString(term) > MaxSearchLength {
		runes := []rune(term)
		term = strings.TrimSpace(string(runes[:MaxSearchLength]))
	}
	return term
}

// MeetsMinLength reports whether term has at least min runes. Use it to skip a
// costly fuzzy search: a pg_trgm index cannot serve terms under 3 characters,
// so callers can fall back to a prefix or exact match for short terms.
func MeetsMinLength(term string, min int) bool {
	return utf8.RuneCountInString(term) >= min
}

// EscapeLike escapes the LIKE/ILIKE wildcards (%, _) and the escape character
// itself, so a user's term matches literally. Ent's ContainsFold, HasPrefix and
// similar predicates already escape; use this only in hand-written SQL.
// The default PostgreSQL escape character is the backslash, so no ESCAPE clause is needed.
func EscapeLike(s string) string {
	return likeEscaper.Replace(s)
}

// ContainsPattern returns "%term%" with the term escaped, for hand-written ILIKE.
func ContainsPattern(term string) string {
	return "%" + EscapeLike(term) + "%"
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
