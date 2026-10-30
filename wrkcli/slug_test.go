package wrkcli

import (
	"strings"
	"testing"
)

func TestSlugifyKeepsASCIIOnly(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ascii words", "Fix login bug", "fix-login-bug"},
		{"punctuation", "Fix Login BUG!!", "fix-login-bug"},
		{"mixed chinese and url", "现在修复http-127-0-0-1-8422页面", "http-127-0-0-1-8422"},
		{"chinese between ascii", "Fix 登录 bug", "fix-bug"},
		{"emoji", "修复 🐛 bug", "bug"},
		{"pure chinese", "修复登录页面", ""},
		{"symbols only", "!!!", ""},
		{"whitespace only", "   ", ""},
		{"empty", "", ""},
		{"accented latin", "CAFÉ naïve", "caf-na-ve"},
		{"collapse dashes", "a--b", "a-b"},
		{"leading trailing dashes", "--a-b--", "a-b"},
		{"ascii digits", "v2 release 42", "v2-release-42"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugify(tc.in)
			if got != tc.want {
				t.Fatalf("slugify(%q) = %q, want %q", tc.in, got, tc.want)
			}
			for _, r := range got {
				if r > 127 {
					t.Fatalf("slugify(%q) = %q contains non-ASCII rune %q", tc.in, got, r)
				}
			}
		})
	}
}

func TestSlugifyTruncatesTo64Runes(t *testing.T) {
	got := slugify(strings.Repeat("a", 70))
	if got != strings.Repeat("a", 64) {
		t.Fatalf("len=%d want 64", len([]rune(got)))
	}
	// Trailing dash from a dropped non-ASCII char is trimmed after truncation.
	got = slugify(strings.Repeat("a", 63) + "修")
	if got != strings.Repeat("a", 63) {
		t.Fatalf("slugify with trailing chinese = %q, want 63 a's", got)
	}
}

func TestNormalizeTaskSlugNeverErrorsOnFreeText(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"   ", ""},
		{"fix login bug", "fix-login-bug"},
		{"修复登录页面", ""},
		{"现在修复http页面", "http"},
		{"!!!", ""},
	}
	for _, tc := range cases {
		got, err := NormalizeTaskSlug(tc.in)
		if err != nil {
			t.Fatalf("NormalizeTaskSlug(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeTaskSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
