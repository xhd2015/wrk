package wrkcli

import (
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func TestFormatWrkStatusFiveBuckets(t *testing.T) {
	got := formatWrkStatus(status.WrkCounts{Untracked: 1})
	want := "dirty (0 staged, 0 changed, 0 renamed, 0 deleted, 1 untracked)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if formatWrkStatus(status.WrkCounts{}) != "clean" {
		t.Fatalf("zero counts should be clean")
	}
}

func TestFormatStatusCountSegmentStagedGreen(t *testing.T) {
	got := formatStatusCountSegment(2, "staged")
	if !strings.Contains(got, ansiGreen) {
		t.Fatalf("staged>0 should be green: %q", got)
	}
	if !strings.Contains(got, "2 staged") {
		t.Fatalf("missing label: %q", got)
	}
	untracked := formatStatusCountSegment(3, "untracked")
	if !strings.Contains(untracked, ansiRed) {
		t.Fatalf("untracked>0 should be red: %q", untracked)
	}
	zero := formatStatusCountSegment(0, "staged")
	if !strings.Contains(zero, ansiGrey) {
		t.Fatalf("zero should be grey: %q", zero)
	}
}

func TestParsePorcelainWrkStagedPriority(t *testing.T) {
	// Fully staged modifications/renames/new files → staged only.
	c := status.ParsePorcelainWrk("M  a.go\nM  b.go\nA  new.go\nR  old.go -> renamed.go\nD  gone.go")
	if c.Staged != 5 || c.Changed != 0 || c.Renamed != 0 || c.Deleted != 0 {
		t.Fatalf("staged-only set: %+v", c)
	}
	// Unstaged-only dirt stays in changed/deleted; ?? → untracked.
	c = status.ParsePorcelainWrk(" M mod.go\n D del.go\n?? u.txt")
	if c.Staged != 0 || c.Changed != 1 || c.Deleted != 1 || c.Untracked != 1 {
		t.Fatalf("unstaged set: %+v", c)
	}
	// AM is staged once, not also changed.
	c = status.ParsePorcelainWrk("AM edit.go")
	if c.Staged != 1 || c.Changed != 0 {
		t.Fatalf("AM: %+v", c)
	}
}

func TestParsePorcelainExactDirtyCountsShape(t *testing.T) {
	p := " M README.md\nA  added.txt\n D delete-me.txt\nR  rename-me.txt -> renamed.txt\n"
	c := status.ParsePorcelainWrk(p)
	if c.Staged != 2 || c.Changed != 1 || c.Deleted != 1 || c.Renamed != 0 || c.Untracked != 0 {
		t.Fatalf("got %+v", c)
	}
}