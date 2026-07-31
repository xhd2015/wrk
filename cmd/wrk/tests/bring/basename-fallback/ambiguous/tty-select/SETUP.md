# Scenario

**Feature**: ambiguous --bring basename resolved once via preflight Select (no double prompt)

```
aaa/mydep + zzz/mydep saved (both example.com/dep)
# preflight resolve-once: one Select + one stdin line; apply reuses abs path
WRK_BASENAME_CONFIRM=1 + stdin "2\n" -> --bring succeeds from zzz/mydep (lex-sorted #2)
```

## Steps

1. Create consumer git repo requiring `example.com/dep`.
2. Create git dep repos `{WorkRoot}/aaa/mydep` and `{WorkRoot}/zzz/mydep`.
3. Record both with `wrk --add`.
4. Run `wrk --bring mydep` with `WRK_BASENAME_CONFIRM=1` and stdin `2\n` (single selection line).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// L2: Capture + StdinInput + BasenameEnv (same harness as other basename-fallback leaves).
	req.InProcess = true
	consumer := initConsumerForBringBasename(t, req.WorkRoot)
	depA := initSavedDepRepo(t, req.WorkRoot, "aaa", "mydep")
	depZ := initSavedDepRepo(t, req.WorkRoot, "zzz", "mydep")
	recordSavedProject(t, req, depA)
	recordSavedProject(t, req, depZ)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = depA
	req.SecondRepo = depZ
	req.SelectedSavedRepo = depZ
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", "mydep"}
	req.BasenameEnv = "WRK_BASENAME_CONFIRM=1"
	// Exactly one stdin line: preflight must not re-prompt in apply.
	req.StdinInput = "2\n"
	return nil
}
```
