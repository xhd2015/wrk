package wrkcli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffDepReplaceUndoActions(t *testing.T) {
	modDir := t.TempDir()
	before := []byte(`module example.com/app

go 1.25

replace example.com/kept => ./kept
replace example.com/changed => ./old
replace example.com/removed => ./removed
`)
	after := []byte(`module example.com/app

go 1.25

replace example.com/kept => ./kept
replace example.com/changed => ./new
replace example.com/added => ../added
replace example.com/remote => example.com/replacement v1.2.3
`)

	drops, err := diffDepReplaceUndoActions(before, after, depUpdateConsumer{
		Path:   "example.com/app",
		ModDir: modDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(drops) != 1 {
		t.Fatalf("got %d drops, want 1: %#v", len(drops), drops)
	}
	got := drops[0]
	if got.modulePath != "example.com/added" {
		t.Fatalf("module path = %q, want %q", got.modulePath, "example.com/added")
	}
	if got.oldSpec != "example.com/added" {
		t.Fatalf("old spec = %q, want %q", got.oldSpec, "example.com/added")
	}
	if got.newPath != "../added" {
		t.Fatalf("new path = %q, want %q", got.newPath, "../added")
	}
}

func TestDiffDepReplaceUndoActionsVersionQualified(t *testing.T) {
	before := []byte("module example.com/app\n\ngo 1.25\n")
	after := []byte("module example.com/app\n\ngo 1.25\n\nreplace example.com/dep v1.2.3 => ../dep\n")

	drops, err := diffDepReplaceUndoActions(before, after, depUpdateConsumer{
		Path:   "example.com/app",
		ModDir: t.TempDir(),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(drops) != 1 {
		t.Fatalf("got %d drops, want 1: %#v", len(drops), drops)
	}
	if drops[0].modulePath != "example.com/dep" || drops[0].oldSpec != "example.com/dep@v1.2.3" {
		t.Fatalf("got drop %#v, want dep@v1.2.3", drops[0])
	}
}

func TestBuildDepReplaceUndoTreeSkipsMissingHeadBaseline(t *testing.T) {
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	goMod := filepath.Join(root, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/app\n\ngo 1.25\n\nreplace example.com/dep => ../dep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tree, warnings, err := buildDepReplaceUndoTree(root, []depUpdateConsumer{
		{Path: "example.com/app", ModDir: root, Checkout: root},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 0 {
		t.Fatalf("got undo tree %#v, want no actions", tree)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "without HEAD baseline") {
		t.Fatalf("got warnings %#v, want missing HEAD baseline warning", warnings)
	}
}

func TestDiffDepReplaceUndoActionsFilter(t *testing.T) {
	before := []byte("module example.com/app\n\ngo 1.25\n")
	after := []byte("module example.com/app\n\ngo 1.25\n\nreplace example.com/one => ../one\nreplace example.com/two => ../two\n")

	drops, err := diffDepReplaceUndoActions(before, after, depUpdateConsumer{
		Path:   "example.com/app",
		ModDir: filepath.Join(t.TempDir(), "app"),
	}, map[string]struct{}{"example.com/two": {}})
	if err != nil {
		t.Fatal(err)
	}
	if len(drops) != 1 || drops[0].modulePath != "example.com/two" {
		t.Fatalf("got drops %#v, want only example.com/two", drops)
	}
}
