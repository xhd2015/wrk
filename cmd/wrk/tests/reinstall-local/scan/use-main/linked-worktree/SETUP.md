# Scenario

**Feature**: useMain from linked worktree resolves scan root to main path (R3)

```
# R3: main repo with root + tools modules; linked worktree sibling checkout
# workDir = linked worktree root; useMain=true
mainrepo/ (git main)  +  linked-wt/ (git worktree add)
workDir=linked-wt + useMain=true
  -> ResolveReinstallScanRoot -> mainrepo (NOT linked-wt)
  -> Plan modules under main: rootbin + toolbin
```

## Steps

1. Init main repo at `{WorkRoot}/mainrepo` with root module
   `example.com/main-root` (`./cmd/rootbin`) and nested `tools`
   `example.com/main-tools` (`./cmd/toolbin`).
2. Commit on main.
3. Add linked worktree at `{WorkRoot}/linked-wt` on branch `side`.
4. Touch `$binDir/rootbin` and `$binDir/toolbin`.
5. Set WorkDir to linked worktree path; UseMain already true from group.
6. Expect WantScanRoot=mainrepo (resolved); assert ScanRoot != linked path;
   Modules both under mainrepo paths.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "mainrepo")
	initGitRepoOnMain(t, mainRepo)

	writeGoMod(t, mainRepo, "example.com/main-root")
	writePackageMain(t, filepath.Join(mainRepo, "cmd", "rootbin"))

	toolsMod := filepath.Join(mainRepo, "tools")
	writeGoMod(t, toolsMod, "example.com/main-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	gitCommitAll(t, mainRepo, "init main multi-module")

	linkedWT := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "side", linkedWT)

	mainRepo = resolvePath(t, mainRepo)
	toolsMod = resolvePath(t, toolsMod)
	linkedWT = resolvePath(t, linkedWT)

	touchBin(t, req.BinDir, "rootbin")
	touchBin(t, req.BinDir, "toolbin")

	req.WorkDir = linkedWT
	req.WantScanRoot = mainRepo
	req.WantModules = []WantModulePlan{
		{
			ModuleRoot: mainRepo,
			ModuleName: "main-root",
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
			ModuleName: "main-tools",
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
