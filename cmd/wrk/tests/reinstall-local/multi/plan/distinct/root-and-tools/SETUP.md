# Scenario

**Feature**: root + nested tools modules both install distinct bins (M2)

```
# M2: {WorkRoot}/root (cmd/rootbin) + {WorkRoot}/root/tools (cmd/toolbin)
# bins present in shared GOBIN; ModuleRoots passed tools-first to prove re-sort
PlanLocalReinstallsMulti([tools, root], binDir)
  -> Modules sorted: root then root/tools
  -> root: install rootbin; tools: install toolbin
```

## Steps

1. Create module `{WorkRoot}/root` path `example.com/multi-root` with
   `./cmd/rootbin` package main.
2. Create nested module `{WorkRoot}/root/tools` path `example.com/multi-tools`
   with `./cmd/toolbin` package main.
3. Touch `$binDir/rootbin` and `$binDir/toolbin`.
4. Pass ModuleRoots as **[tools, root]** (reverse lex order) so product must
   re-sort by ModuleRoot path.
5. Expect Modules ordered root < root/tools with one install item each.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	rootMod := filepath.Join(req.WorkRoot, "root")
	toolsMod := filepath.Join(rootMod, "tools")

	writeGoMod(t, rootMod, "example.com/multi-root")
	writePackageMain(t, filepath.Join(rootMod, "cmd", "rootbin"))

	writeGoMod(t, toolsMod, "example.com/multi-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	touchBin(t, req.BinDir, "rootbin")
	touchBin(t, req.BinDir, "toolbin")

	// Caller order intentionally reverse of lex ModuleRoot order.
	req.ModuleRoots = []string{toolsMod, rootMod}
	req.WantModules = []WantModulePlan{
		{
			ModuleRoot: rootMod,
			ModuleName: "multi-root",
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
			ModuleName: "multi-tools",
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
