# Scenario

**Feature**: git multi-module checkout; workDir=subdir still plans all modules (R1)

```
# R1: repo with root go.mod + nested tools/go.mod; workDir=pkg/sub (no go.mod)
# useMain=false → scanRoot = git toplevel (repo), not walk-up-only single module
repo/ (git)
  go.mod + cmd/rootbin
  tools/go.mod + cmd/toolbin
  pkg/sub/   <- workDir
  -> ResolveReinstallScanRoot -> repo
  -> PlanLocalReinstallsFromWorkDir
  -> Modules: root (rootbin) + tools (toolbin), both install
```

## Steps

1. Init git repo at `{WorkRoot}/repo` on branch `main`.
2. Write root module `example.com/scan-root` with `./cmd/rootbin` package main.
3. Write nested module `{repo}/tools` `example.com/scan-tools` with
   `./cmd/toolbin` package main.
4. Create empty subdirectory `{repo}/pkg/sub` (no go.mod) as WorkDir.
5. Commit tree so ShowToplevel is valid.
6. Touch `$binDir/rootbin` and `$binDir/toolbin`.
7. Expect ScanRoot=repo; Modules lex: repo then repo/tools, one install each.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepoOnMain(t, repo)

	writeGoMod(t, repo, "example.com/scan-root")
	writePackageMain(t, filepath.Join(repo, "cmd", "rootbin"))

	toolsMod := filepath.Join(repo, "tools")
	writeGoMod(t, toolsMod, "example.com/scan-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	sub := filepath.Join(repo, "pkg", "sub")
	mkdirAll(t, sub)

	gitCommitAll(t, repo, "init multi-module checkout")

	repo = resolvePath(t, repo)
	toolsMod = resolvePath(t, toolsMod)
	sub = resolvePath(t, sub)

	touchBin(t, req.BinDir, "rootbin")
	touchBin(t, req.BinDir, "toolbin")

	req.WorkDir = sub
	req.WantScanRoot = repo
	req.WantModules = []WantModulePlan{
		{
			ModuleRoot: repo,
			ModuleName: "scan-root",
			Items: []WantPlanItem{
				{
					BinName: "rootbin",
					Method:  methodGoInstall,
					RelPath: "./cmd/rootbin",
					Action:  actionInstall,
				},
			},
		},
		{
			ModuleRoot: toolsMod,
			ModuleName: "scan-tools",
			Items: []WantPlanItem{
				{
					BinName: "toolbin",
					Method:  methodGoInstall,
					RelPath: "./cmd/toolbin",
					Action:  actionInstall,
				},
			},
		},
	}
	return nil
}
```
