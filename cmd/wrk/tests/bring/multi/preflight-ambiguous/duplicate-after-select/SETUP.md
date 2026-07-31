# Scenario

**Feature**: multi-bring rejects when basename selection resolves to the same abs path as another --bring

```
# aaa/mydep + zzz/mydep saved; second arg is abs path of zzz/mydep
# WRK_BASENAME_CONFIRM=1 + stdin "2\n" selects zzz/mydep for basename
#   -> preflight: one Select, then duplicate resolved abs
#   -> non-zero; wrk: duplicate --bring path: <abs>
#   -> no will bring plan; no external/ create; empty stdout preferred
consumer requires dep1
  -> --bring mydep --bring <zzz/mydep abs>
  -> Select 2 -> duplicate error
```

## Steps

1. Create consumer requiring `example.com/dep1`.
2. Create and `wrk --add` `aaa/mydep` and `zzz/mydep` (same module).
3. Run `wrk --bring mydep --bring <abs-of-zzz/mydep>` with `WRK_BASENAME_CONFIRM=1` and stdin `2\n`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true

	// Single-module consumer (only dep1) — second arg is the same dep path, not a second module.
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runBringGo(t, consumer, "mod", "edit", "-require="+multiBringDep1Module+"@v0.0.0")
	var err error
	consumer, err = filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks consumer: %v", err)
	}

	mydepA := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("aaa", "mydep"), multiBringDep1Module)
	mydepZ := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("zzz", "mydep"), multiBringDep1Module)
	multiPreflightRecordSaved(t, req, mydepA)
	multiPreflightRecordSaved(t, req, mydepZ)

	// Basename first (needs Select); second arg is abs of the selected candidate.
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = mydepA
	req.SecondRepo = mydepZ
	req.SelectedSavedRepo = mydepZ // stdin 2 → lex #2 = zzz/mydep
	req.Args = []string{"--bring", "mydep", "--bring", mydepZ}
	req.BasenameEnv = "WRK_BASENAME_CONFIRM=1"
	req.StdinInput = "2\n"
	return nil
}
```
