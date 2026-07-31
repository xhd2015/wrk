# Scenario

**Feature**: multi-bring confirms two ambiguous basenames one-by-one then prints will bring plan

```
# aaa/mydep + zzz/mydep; aaa/otherlib + zzz/otherlib saved
# WRK_BASENAME_CONFIRM=1 + stdin "2\n1\n"
#   -> Select mydep (#2 zzz) then otherlib (#1 aaa)
#   -> stderr will bring: mydep → zzz/mydep; otherlib → aaa/otherlib
#   -> stdout two external paths; both worktrees + replaces
consumer requires dep1+dep2
  -> multi preflight Select left→right
  -> will bring plan (multi-only)
  -> apply
```

## Steps

1. Create consumer requiring `example.com/dep1` and `example.com/dep2`.
2. Create and `wrk --add` four saved deps:
   - `aaa/mydep`, `zzz/mydep` (dep1 module)
   - `aaa/otherlib`, `zzz/otherlib` (dep2 module)
3. Run `wrk --bring mydep --bring otherlib` with `WRK_BASENAME_CONFIRM=1` and stdin `2\n1\n`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initMultiBringConsumerWithTwoRequires(t, req.WorkRoot)

	// Nested parents so filepath.Base is mydep / otherlib for projects.json match.
	mydepA := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("aaa", "mydep"), multiBringDep1Module)
	mydepZ := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("zzz", "mydep"), multiBringDep1Module)
	otherA := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("aaa", "otherlib"), multiBringDep2Module)
	otherZ := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("zzz", "otherlib"), multiBringDep2Module)

	multiPreflightRecordSaved(t, req, mydepA)
	multiPreflightRecordSaved(t, req, mydepZ)
	multiPreflightRecordSaved(t, req, otherA)
	multiPreflightRecordSaved(t, req, otherZ)

	// stdin 2\n1\n → mydep picks lex #2 (zzz), otherlib picks lex #1 (aaa).
	// Assert rebuilds all four paths from WorkRoot layout (aaa/zzz + basename).
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = mydepZ    // selected for mydep
	req.SecondRepo = otherA // selected for otherlib
	req.Args = []string{"--bring", "mydep", "--bring", "otherlib"}
	req.BasenameEnv = "WRK_BASENAME_CONFIRM=1"
	req.StdinInput = "2\n1\n"
	return nil
}
```