package wrkcli

import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

func TestConsumerReplaceAlreadyEquivalent(t *testing.T) {
	root := t.TempDir()
	depDir := root // module at checkout root
	modDir := filepath.Join(root, "cmd")
	absDep := storage.NormalizePath(depDir)

	cRel := depUpdateConsumer{
		Path:   "example.com/dep/cmd",
		ModDir: modDir,
		Replaces: []scan.ModuleReplace{
			{OldPath: "example.com/dep", NewPath: "../"},
		},
	}
	if !consumerReplaceAlreadyEquivalent(cRel, "example.com/dep", absDep) {
		t.Fatalf("relative ../ from cmd/ should be equivalent to dep root %s", absDep)
	}

	cAbs := depUpdateConsumer{
		Path:   "example.com/app",
		ModDir: filepath.Join(root, "app"),
		Replaces: []scan.ModuleReplace{
			{OldPath: "example.com/dep", NewPath: absDep},
		},
	}
	if !consumerReplaceAlreadyEquivalent(cAbs, "example.com/dep", absDep) {
		t.Fatalf("absolute NewPath equal to absDir should be equivalent")
	}

	cOther := depUpdateConsumer{
		Path:   "example.com/app",
		ModDir: filepath.Join(root, "app"),
		Replaces: []scan.ModuleReplace{
			{OldPath: "example.com/dep", NewPath: filepath.Join(root, "other")},
		},
	}
	if consumerReplaceAlreadyEquivalent(cOther, "example.com/dep", absDep) {
		t.Fatalf("different NewPath must not be equivalent")
	}

	cNone := depUpdateConsumer{
		Path:   "example.com/app",
		ModDir: filepath.Join(root, "app"),
	}
	if consumerReplaceAlreadyEquivalent(cNone, "example.com/dep", absDep) {
		t.Fatalf("no replace must not be equivalent")
	}
}
