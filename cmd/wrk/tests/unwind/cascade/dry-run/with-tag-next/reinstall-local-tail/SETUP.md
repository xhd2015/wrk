# Scenario

**Feature**: cascade with `--reinstall-local` still shows reinstall tail (C-DR6)

```
# multi-repo both dirty + tag-next + reinstall-local
  -> peels free-first
  -> cascade would: tag-next / would: pin
  -> would: reinstall local binaries
  -> exit 0; zero mutations (no reinstall executes)
```

## Steps

1. Seed multi-repo both-dirty cascade fixture.
2. Run dry-run with tag-next/push/done **and** `--reinstall-local`.
3. Expect cascade present and reinstall tail line; no mutations.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeMultiRepoBothDirty(t, req)
	req.Args = []string{
		"--unwind", "--dry-run",
		"--tag-next", "--push", "--done",
		"--reinstall-local",
	}
	recordUnwindBaseline(t, req)
	return nil
}
```
