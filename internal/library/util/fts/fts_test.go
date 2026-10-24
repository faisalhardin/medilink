package fts

import (
	"testing"
)

func TestSanitizeTsQuery(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single word",
			input: "append",
			want:  "append:*",
		},
		{
			name:  "partial word triggers prefix",
			input: "laparoscop",
			want:  "laparoscop:*",
		},
		{
			name:  "multi-word becomes AND terms",
			input: "knee arthro",
			want:  "knee:* & arthro:*",
		},
		{
			name:  "three words",
			input: "open heart surgery",
			want:  "open:* & heart:* & surgery:*",
		},
		{
			name:  "strips ampersand",
			input: "nerve & block",
			want:  "nerve:* & block:*",
		},
		{
			name:  "strips pipe",
			input: "nerve | block",
			want:  "nerve:* & block:*",
		},
		{
			name:  "strips exclamation",
			input: "!heart",
			want:  "heart:*",
		},
		{
			name:  "strips parentheses — splits into separate tokens",
			input: "lapar(oscopy)",
			want:  "lapar:* & oscopy:*",
		},
		{
			name:  "strips angle brackets and lone hyphen",
			input: "left <-> knee",
			want:  "left:* & knee:*",
		},
		{
			name:  "strips colon",
			input: "code:86",
			want:  "code:* & 86:*",
		},
		{
			name:  "strips single quote — drops lone single-char token",
			input: "doctor's",
			want:  "doctor:*",
		},
		{
			name:  "strips backslash",
			input: `knee\arthro`,
			want:  "knee:* & arthro:*",
		},
		{
			name:  "extra whitespace collapsed",
			input: "  knee   replacement  ",
			want:  "knee:* & replacement:*",
		},
		{
			name:  "only special chars returns empty",
			input: "!&|()",
			want:  "",
		},
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "code prefix with dot",
			input: "86.0",
			want:  "86.0:*",
		},
		{
			name:  "ICD code space description",
			input: "86 appendectomy",
			want:  "86:* & appendectomy:*",
		},
		{
			name:  "partial Indonesian word kari (for karies)",
			input: "kari",
			want:  "kari:*",
		},
		{
			name:  "partial word pulpi (for pulpitis)",
			input: "pulpi",
			want:  "pulpi:*",
		},
		{
			name:  "Indonesian phrase gigi berlubang",
			input: "gigi berlubang",
			want:  "gigi:* & berlubang:*",
		},
		{
			name:  "Indonesian partial radang",
			input: "radang",
			want:  "radang:*",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeTsQuery(tc.input)
			if got != tc.want {
				t.Errorf("SanitizeTsQuery(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
