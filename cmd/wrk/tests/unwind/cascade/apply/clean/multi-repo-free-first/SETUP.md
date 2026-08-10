# Scenario

**Feature**: multi-repo free-first apply cascade (C-AP2)

```
# root requires leaf@v0.0.1; both dirty; leaf owned-changed → next v0.0.2
leaf ← root (repos + modules)
  -> wrk --unwind --tag-next --push --done
  -> land free-first: peel linked leaf then primary
  -> cascade: tag leaf module @ v0.0.2 (+ push when leaf free of pending)
  -> pin root <- leaf @ v0.0.2; cascade pin commit on root
  -> exit 0; leaf tagged/pushed before consumer pin completion
```

## Steps

1. Seed multi-repo apply cascade fixture (both dirty; bare origins; modproxy).
2. Run `--unwind --tag-next --push --done`.
3. Assert free-first module order across repos + cascade pin commit on root.

## Context

- Cross-repo edges → `--tag-next` + `--push`; linked leaf → `--done`.
- Keep-replace is single-repo only (`--done` may remove nested external).
- **RED** while peel still runs TagNextAll + repo-edge pin without cascade pin
  commit that leaves go.mod clean.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeMultiRepoBothDirty(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push", "--done"}
	return nil
}
```
