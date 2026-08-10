# Scenario

**Feature**: free release must be **pushed** before cross-repo cascade pin
drops replace and runs `go mod tidy` against the new version

```
# free monorepo (root + nested same-main pin) + external consumer (earlier alpha path)
# remainingTouchesMain must not hold free push past network pin of consumer
stack free nested + cross-repo consumer + --unwind --tag-next --push
  -> free root tag-next → free push → then pin consumer @ next
  -> exit 0; no unknown revision
```

## Preconditions

- Parent `cascade/apply/` helpers: `setupApplyCascadePushBeforeCrossRepoPin`,
  `assertFreePushBeforeCrossRepoPinOfFree`.
- Offline `file://` modproxy still seeds free@next so tidy can succeed when
  push is late — the **RED surface is order** (push before cross-repo pin),
  matching production where late push → `unknown revision` on real VCS.

## Steps

1. Grouping scopes push-before-cross-repo-pin leaves.
2. Leaves fix Kahn/module-path shapes that defer free push.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```
