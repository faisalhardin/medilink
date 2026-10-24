package fts

import (
	"strings"
	"unicode"
)

// SanitizeTsQuery converts raw user input into a safe PostgreSQL to_tsquery
// expression with prefix matching on every token.
//
// Steps:
//  1. Strip characters that have special meaning in tsquery syntax.
//  2. Collapse whitespace and split into tokens.
//  3. Drop tokens shorter than 2 characters (noise like lone "s").
//  4. Join tokens with " & " (AND) and append ":*" for prefix matching.
//
// Examples:
//
//	"append"        → "append:*"
//	"knee arthro"   → "knee:* & arthro:*"
//	"lapar(oscopy)" → "lapar:* & oscopy:*"
//	"doctor's"      → "doctor:*"          (lone "s" is filtered out)
//	""              → ""                  (caller should guard against empty)
//	"kari"          → "kari:*"            (partial word for prefix matching)
func SanitizeTsQuery(input string) string {
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		switch r {
		case '&', '|', '!', '(', ')', '<', '>', ':', '\'', '\\', '-':
			b.WriteRune(' ')
		default:
			if unicode.IsPrint(r) {
				b.WriteRune(r)
			}
		}
	}

	parts := strings.Fields(b.String())
	terms := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) >= 2 {
			terms = append(terms, p+":*")
		}
	}

	if len(terms) == 0 {
		return ""
	}
	return strings.Join(terms, " & ")
}
