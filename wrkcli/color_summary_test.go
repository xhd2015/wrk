package wrkcli

import (
	"strings"
	"testing"
)

func TestFormatSyncSummaryLinePlain(t *testing.T) {
	got := formatSyncSummaryLine(0, 4, 3, false)
	want := "synced: 0 into main, 4 into worktrees, 3 skipped"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatSyncSummaryLineColor(t *testing.T) {
	got := formatSyncSummaryLine(0, 4, 3, true)
	if strings.Contains(got, "\x1b[") == false {
		t.Fatalf("expected ANSI in colored sync summary: %q", got)
	}
	// Body still carries the numbers and labels (strip check via contains).
	for _, frag := range []string{"into main", "into worktrees", "skipped", "4", "3", "0"} {
		if !strings.Contains(got, frag) {
			t.Fatalf("missing %q in %q", frag, got)
		}
	}
	// Skipped >0 uses orange; worktrees >0 uses green.
	if !strings.Contains(got, ansiOrange) {
		t.Fatalf("expected orange for skipped>0: %q", got)
	}
	if !strings.Contains(got, ansiGreen) {
		t.Fatalf("expected green for into-worktrees>0: %q", got)
	}
}

func TestFormatGoInstallProgressLineHighlight(t *testing.T) {
	plain := formatGoInstallProgressLine(MethodGoInstall, "./cmd/wrk", false)
	if plain != "go install ./cmd/wrk" {
		t.Fatalf("plain: got %q", plain)
	}
	colored := formatGoInstallProgressLine(MethodGoInstall, "./cmd/wrk", true)
	if !strings.HasPrefix(colored, ansiGreen+"go install"+ansiReset) {
		t.Fatalf("colored verb missing green: %q", colored)
	}
	if !strings.HasSuffix(colored, " ./cmd/wrk") {
		t.Fatalf("path not plain suffix: %q", colored)
	}
}

func TestFormatReinstallSummaryLinePlain(t *testing.T) {
	got := formatReinstallSummaryLine(1, 0, 0, false)
	want := "reinstalled 1, skipped 0, failed 0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatUnwindSummaryLineStages(t *testing.T) {
	flags := UnwindFlags{TagNext: true, Push: true, ReinstallLocal: true}
	stats := UnwindApplyStats{HadPeels: true, Peeled: 1, Tagged: 1, Pinned: 0, Pushed: 1, Reinstalled: 1}
	got := formatUnwindSummaryLine(stats, flags, false)
	want := "unwind: peeled 1, tagged 1, pinned 0, pushed 1, reinstalled 1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// Omit stages not requested.
	flags2 := UnwindFlags{ReinstallLocal: true}
	stats2 := UnwindApplyStats{HadPeels: false, Reinstalled: 1}
	got2 := formatUnwindSummaryLine(stats2, flags2, false)
	if got2 != "unwind: reinstalled 1" {
		t.Fatalf("got %q", got2)
	}
	if formatUnwindSummaryLine(UnwindApplyStats{}, UnwindFlags{}, false) != "" {
		t.Fatal("empty stats should yield empty summary")
	}
}
