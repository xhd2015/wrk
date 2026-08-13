# Scenario

**Bug / gap**: `--add-all` feature gen-commit must keep pin `go.sum` hashes (GS1)

```
# T1 + replace-tidy already dropped network hashes for free @ v0.0.1
# pin+tidy on Base writes those hashes; partial-edit restores hashless go.sum
# --add-all gen-commit must not scoop that restore over the pin commit
# (crime scene: cbea1ecf deleted agent-pro v0.0.123 go.sum lines)
root-linked (FEATURE_WIP + uncommitted hashless go.sum + committed replace)
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next
  -> pin commit go.sum has example.com/dot-pkgs v0.0.1 hashes
  -> feature commit tree still has those hashes
  -> HEAD go.sum still has those hashes
  -> exit 0
```

## Steps

1. Seed GS1 (`setupApplyPinBeforeFeatureAddAllKeepPinGoSum`): T1 stack, then
   `go mod tidy` on the linked consumer while the external replace is present
   so `go.sum` drops `example.com/dot-pkgs v0.0.1` hashes; leave `go.sum`
   uncommitted; `FEATURE_WIP.md` already staged.
2. Run apply with gen-commit + `--add-all` + `--merge-back` + `--tag-next`
   (same flag set as T1).
3. Assert pin `go.sum` has hashes; feature tree and HEAD still have them.

## Context

- Distinct from T1 (no hashless `go.sum` dirt) and from
  `dirty-gomod/with-add-all` (no feature peel after restore).
- Distinct from CS-repin (no deferred free tag / require `@ next`).
- `assertGoModCommittedClean` is porcelain-only and would miss this hole.
- Classic TDD: expect **RED** until `--add-all` gen-commit does not restore
  the pre-pin hashless `go.sum` over the pin commit.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureAddAllKeepPinGoSum(t, req)
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
	)
	return nil
}
```
