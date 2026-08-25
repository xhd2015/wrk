# Scenario

**Feature**: nested intra-repo relative replace inside the dep checkout is left alone

```
# primary requires dep (no replace); external/dep is nested git on the stack
# dep/cmd has replace example.com/dep => ../
cwd=primary -> wrk --dep-replace <external/dep>
  -> primary gains absolute replace
  -> dep root skipped (self)
  -> dep/cmd keeps => ../
```

## Steps

1. Seed primary + on-stack multi-module dep with nested `cmd` relative replace.
2. Run apply from primary.

```go
import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

const modDepCmd = "example.com/dep/cmd"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true

	primary := initStackPrimary(t, req)
	dep := seedExternalGitModule(t, primary, "dep", modDep, "")
	writeLibPkg(t, dep, "dep", "Version")

	cmdDir := filepath.Join(dep, "cmd")
	writeGoMod(t, cmdDir, modDepCmd,
		fmt.Sprintf("require %s v0.0.1\n\nreplace %s => ../\n", modDep, modDep))
	writeLibPkg(t, cmdDir, "cmd", "Main")
	gitCommitAll(t, dep, "dep + cmd relative replace")
	dep = resolvePath(t, dep)
	req.DepDir = dep

	body := fmt.Sprintf("require %s v0.0.1\n", modDep)
	writeGoMod(t, primary, modApp, body)
	writeConsumerMainWithImports(t, primary, modDep)
	gitCommitAll(t, primary, "primary requires dep")

	primary = resolvePath(t, primary)
	req.RepoDir = primary
	req.ConsumerModDir = primary
	req.ConsumerGoMod = filepath.Join(primary, "go.mod")
	req.BaselineGoMod = readFile(t, req.ConsumerGoMod)
	req.WantConsumerModule = modApp

	cmdDir = resolvePath(t, cmdDir)
	req.Consumer2ModDir = cmdDir
	req.Consumer2GoMod = filepath.Join(cmdDir, "go.mod")
	req.Baseline2GoMod = readFile(t, req.Consumer2GoMod)
	req.WantConsumer2Module = modDepCmd
	req.WantCheckout = "."
	req.WantUpdated = 1
	req.WantCheckouts = 1

	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
