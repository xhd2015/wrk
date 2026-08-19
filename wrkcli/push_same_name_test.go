package wrkcli

import (
	"strings"
	"testing"
)

func TestParseLsRemoteHeadSHA(t *testing.T) {
	t.Parallel()
	const sha = "0123456789abcdef0123456789abcdef01234567"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace", "  \n", ""},
		{"tab ref", sha + "\trefs/heads/feature\n", sha},
		{"spaces", sha + " refs/heads/feature", sha},
		{"first of two", sha + "\trefs/heads/feature\n" +
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\trefs/heads/other\n", sha},
		{"not a sha", "not-a-sha\trefs/heads/feature\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := parseLsRemoteHeadSHA(tc.in); got != tc.want {
				t.Fatalf("parseLsRemoteHeadSHA(%q)=%q want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestDecideSameNameRemoteUpdate(t *testing.T) {
	t.Parallel()
	snap := sameNameRemoteSnapshot{
		branch:       "feat",
		remoteExists: true,
		remoteTip:    "aaa111",
		included:     true,
		localHead:    "bbb222",
	}
	t.Run("no remote", func(t *testing.T) {
		t.Parallel()
		kind, warn := decideSameNameRemoteUpdate(sameNameRemoteSnapshot{}, true, "aaa111")
		if kind != sameNameRemoteSkipNone || warn != "" {
			t.Fatalf("got kind=%d warn=%q", kind, warn)
		}
	})
	t.Run("tip unchanged included", func(t *testing.T) {
		t.Parallel()
		kind, warn := decideSameNameRemoteUpdate(snap, true, "aaa111")
		if kind != sameNameRemoteDoUpdate || warn != "" {
			t.Fatalf("got kind=%d warn=%q", kind, warn)
		}
	})
	t.Run("tip moved", func(t *testing.T) {
		t.Parallel()
		kind, warn := decideSameNameRemoteUpdate(snap, true, "ccc333")
		if kind != sameNameRemoteSkipMoved {
			t.Fatalf("got kind=%d", kind)
		}
		if !containsAll(warn, "origin/feat moved", "aaa111", "ccc333") {
			t.Fatalf("warn=%q", warn)
		}
	})
	t.Run("gone", func(t *testing.T) {
		t.Parallel()
		kind, warn := decideSameNameRemoteUpdate(snap, false, "")
		if kind != sameNameRemoteSkipGone {
			t.Fatalf("got kind=%d", kind)
		}
		if !containsAll(warn, "origin/feat disappeared") {
			t.Fatalf("warn=%q", warn)
		}
	})
	t.Run("not included", func(t *testing.T) {
		t.Parallel()
		s := snap
		s.included = false
		kind, warn := decideSameNameRemoteUpdate(s, true, "aaa111")
		if kind != sameNameRemoteSkipNotIncluded {
			t.Fatalf("got kind=%d", kind)
		}
		if !containsAll(warn, "not in local branch", "not force-updating") {
			t.Fatalf("warn=%q", warn)
		}
	})
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
