# Scenario

**Feature**: TTY named create at a missing exact path under `{WRK_HOME}/worktrees` creates there (no Policy B)

```
# reusable sibling already under the dump; preferred branch main-{date} taken
# TTY + Enter (default Y) would reuse the sibling *today* if Policy B ran
# after skip: create exactly at the missing dump name; no would-reuse / skip-creating
{WRK_HOME}/worktrees/myrepo-main-{date} (reusable)
  -> wrk myrepo {WRK_HOME}/worktrees/myrepo-main-{date}-named  (stdin \n)
  -> stdout = new exact path; sibling still listed
```

## Steps

1. Grouping already added one reusable dump sibling (clean, HEAD==source).
2. Set `req.SpawnDir` to a **new missing** path under `{req.WrkHome}/worktrees`.
3. Run named create under fake TTY with stdin `\n` (default Y).

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseScriptTTY = true
	if !req.UseScriptTTY {
		t.Fatal("UseScriptTTY must be true for dump-parent TTY leaf")
	}
	// Policy B prompt is gated on term.IsTerminal(stdin); script(1) e2e.
	req.InProcess = false

	dump := policyBDumpParent(req)
	sib := req.WtDir
	srcHEAD := revParseHEAD(t, req.TargetDir)
	if got := revParseHEAD(t, sib); got != srcHEAD {
		t.Fatalf("dump sibling HEAD %s != source %s", got, srcHEAD)
	}
	if st := strings.TrimSpace(gitOutputIsolated(t, sib, "status", "--porcelain")); st != "" {
		t.Fatalf("dump sibling not clean: %q", st)
	}

	spawn := filepath.Join(dump, "myrepo-main-"+wrkDate+"-named")
	if filepath.Clean(filepath.Dir(spawn)) != filepath.Clean(dump) {
		t.Fatalf("spawn parent %q != dump %q", filepath.Dir(spawn), dump)
	}
	if spawn == sib {
		t.Fatalf("spawn path must differ from dump sibling %q", sib)
	}
	if _, err := os.Stat(spawn); err == nil {
		t.Fatalf("spawn path must be missing: %s", spawn)
	}
	req.SpawnDir = spawn
	req.StdinInput = "\n"
	return nil
}
```
