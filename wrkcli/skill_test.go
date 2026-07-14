package wrkcli

import (
	"strings"
	"testing"
)

func TestParseSkillShowArgsHeader(t *testing.T) {
	header, err := parseSkillShowArgs([]string{"--header"})
	if err != nil {
		t.Fatalf("parseSkillShowArgs: %v", err)
	}
	if !header {
		t.Fatal("expected header=true")
	}
}

func TestParseSkillShowArgsRejectsUnknownFlag(t *testing.T) {
	_, err := parseSkillShowArgs([]string{"--nope"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "--nope") {
		t.Fatalf("error should mention --nope, got %q", err)
	}
}

func TestFindWrkModeFlag(t *testing.T) {
	flag, found := findWrkModeFlag([]string{"list", "--done"})
	if !found || flag != "--done" {
		t.Fatalf("findWrkModeFlag = (%q, %v), want (--done, true)", flag, found)
	}
}