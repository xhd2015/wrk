package unwind

import "testing"

func TestFormatDepUpdateCommitMsg(t *testing.T) {
	t.Parallel()
	got := FormatDepUpdateCommitMsg([]DepUpdateBump{{
		Module: "github.com/xhd2015/agent-pro", From: "v0.0.192", To: "v0.0.193",
	}})
	want := "dep: github.com/xhd2015/agent-pro v0.0.192 -> v0.0.193"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = FormatDepUpdateCommitMsg([]DepUpdateBump{
		{Module: "github.com/xhd2015/dot-pkgs/go-pkgs", From: "go-pkgs/v0.0.169", To: "go-pkgs/v0.0.170"},
		{Module: "github.com/xhd2015/agent-pro", From: "v0.0.192", To: "v0.0.193"},
	})
	want = "deps: github.com/xhd2015/agent-pro v0.0.192 -> v0.0.193, github.com/xhd2015/dot-pkgs/go-pkgs v0.0.169 -> v0.0.170"
	if got != want {
		t.Fatalf("multi got %q want %q", got, want)
	}
}
