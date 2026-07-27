# Scenario

**Feature**: same BinName skip×skip across modules does not hard-error (M4)

```
# M4 (skip-only duplicate): mod-a and mod-b both ./cmd/sharedbin; bin ABSENT
# → both Action=skip; multi plan ok (only install×install is a hard collision)
# ModuleRoots passed mod-b first; product re-sorts lex by ModuleRoot
PlanLocalReinstallsMulti([mod-b, mod-a], binDir)
  -> Modules=[mod-a skip sharedbin, mod-b skip sharedbin], err=nil
```

## Steps

1. Create `{WorkRoot}/mod-a` module `example.com/mod-a` with `./cmd/sharedbin`.
2. Create `{WorkRoot}/mod-b` module `example.com/mod-b` with `./cmd/sharedbin`.
3. Do **not** create `$binDir/sharedbin` (both Actions=skip under shared binDir).
4. Pass ModuleRoots as `[mod-b, mod-a]` (reverse of lex path order).
5. Expect nil error; two modules lex-sorted, each with one skip item for
   `sharedbin`.

## Context

- With a shared `binDir`, the same BinName always gets the same Action from
  presence filtering, so a constructible non-collision duplicate is
  **skip×skip** (bin absent). That locks exit criterion “skip-only duplicates
  do not hard-error”. install×install is covered under `error/install-install`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	modA := filepath.Join(req.WorkRoot, "mod-a")
	modB := filepath.Join(req.WorkRoot, "mod-b")

	writeGoMod(t, modA, "example.com/mod-a")
	writePackageMain(t, filepath.Join(modA, "cmd", "sharedbin"))

	writeGoMod(t, modB, "example.com/mod-b")
	writePackageMain(t, filepath.Join(modB, "cmd", "sharedbin"))
	// intentionally no touchBin for "sharedbin" → skip on both modules

	req.ModuleRoots = []string{modB, modA}
	req.WantModules = []WantModulePlan{
		{
			ModuleRoot: modA,
			ModuleName: "mod-a",
			Items: []WantPlanItem{
				{
					BinName: "sharedbin",
					Method:  methodGoInstall,
					RelPath: "./cmd/sharedbin",
					Action:  actionSkip,
				},
			},
		},
		{
			ModuleRoot: modB,
			ModuleName: "mod-b",
			Items: []WantPlanItem{
				{
					BinName: "sharedbin",
					Method:  methodGoInstall,
					RelPath: "./cmd/sharedbin",
					Action:  actionSkip,
				},
			},
		},
	}
	return nil
}
```
