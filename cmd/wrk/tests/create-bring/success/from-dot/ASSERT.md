## Expected

- Exit code 0.
- Create path `{WRK_HOME}/worktrees/src-main-2026-06-30` exists as a linked worktree and is the only stdout line (`--no-new-terminal` quiet bring).
- External `mydep1` lives under **that** worktree, not under `src`.
- `src/external` does not exist; `src/go.mod` is unchanged (no replace).
- Last event `command=="create"`; `args` contain `--bring` and the dep path (positionals are not event args).

## Side Effects

- New worktree under `{WRK_HOME}/worktrees`; bring consumer is that path, not `src`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wt := createBringDefaultWT(req)
	ext1 := createBringExternalPath(wt, "mydep1")
	req.ConsumerTop = wt
	req.ExternalWtDir = ext1

	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	assertWorktreeListContains(t, req.MainRepo, wt)
	if strings.TrimSpace(resp.Stdout) != wt {
		t.Fatalf("stdout should be the create path only %q; got %q", wt, resp.Stdout)
	}
	assertFileExists(t, ext1)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "external"))

	mod, err := readCreateBringGoMod(wt)
	if err != nil {
		t.Fatalf("read new WT go.mod: %v", err)
	}
	if !createBringHasReplace(mod, createBringDep1Module, ext1) {
		t.Fatalf("new WT go.mod missing replace %s => %s: %+v", createBringDep1Module, ext1, mod.Replace)
	}

	ev := createBringLastEvent(t, req.WrkHome)
	if ev.Command != "create" {
		t.Fatalf("event command = %q, want create", ev.Command)
	}
	if !createBringArgsContain(ev.Args, "--bring") {
		t.Fatalf("event args should include --bring, got %v", ev.Args)
	}
	if !createBringArgsContain(ev.Args, req.DepPath) {
		t.Fatalf("event args should include dep %q, got %v", req.DepPath, ev.Args)
	}
}
```
