package wrkcli

import (
	"errors"
	"fmt"
	"testing"
)

func TestCompactStackCheckoutErr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "fatal_prefix",
			err:  errors.New("git status --porcelain in /tmp/x: fatal: not a git repository: /root/xgo/.git/worktrees/xgo-dev"),
			want: "not a git repository: /root/xgo/.git/worktrees/xgo-dev",
		},
		{
			name: "git_status_wrapper",
			err:  errors.New("git status: exit status 128"),
			want: "exit status 128",
		},
		{
			name: "plain",
			err:  errors.New("boom"),
			want: "boom",
		},
		{
			name: "nil",
			err:  nil,
			want: "",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := compactStackCheckoutErr(tc.err)
			if got != tc.want {
				t.Fatalf("compactStackCheckoutErr() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatSkipNestedCheckoutWarning(t *testing.T) {
	t.Parallel()
	root := "/tmp/primary"
	nested := "/tmp/primary/sandbox/broken-wt"
	err := fmt.Errorf("git status --porcelain in %s: fatal: not a git repository: /root/xgo/.git/worktrees/xgo-dev", nested)
	got := formatSkipNestedCheckoutWarning(root, nested, err)
	wantSub := "warning: skipping nested checkout sandbox/broken-wt: not a git repository: /root/xgo/.git/worktrees/xgo-dev"
	if got != wantSub {
		t.Fatalf("got %q, want %q", got, wantSub)
	}
}
