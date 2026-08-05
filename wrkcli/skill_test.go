package wrkcli

import (
	"testing"
)

func TestFindWrkModeFlag(t *testing.T) {
	flag, found := findWrkModeFlag([]string{"list", "--done"})
	if !found || flag != "--done" {
		t.Fatalf("findWrkModeFlag = (%q, %v), want (--done, true)", flag, found)
	}
}
