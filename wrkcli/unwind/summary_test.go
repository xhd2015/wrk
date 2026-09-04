package unwind

import (
	"testing"
)

func TestFormatUnwindSummaryLineStages(t *testing.T) {
	flags := UnwindFlags{TagNext: true, Push: true, ReinstallLocal: true}
	stats := UnwindApplyStats{HadPeels: true, Peeled: 1, Tagged: 1, Pinned: 0, Pushed: 1, Reinstalled: 1}
	got := formatUnwindSummaryLine(stats, flags, false)
	want := "unwind: peeled 1, tagged 1, pinned 0, pushed 1, reinstalled 1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
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
