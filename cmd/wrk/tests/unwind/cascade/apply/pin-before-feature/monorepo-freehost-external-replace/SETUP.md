# Scenario

**Feature**: B1 apply — monorepo freeHost + clean external replace pin **before** feature gen-commit (T-M1)

```
# dirty linked monorepo freeHost: pkgs/shared owned-changed (same-label free)
# + root replace => ./external/dot-pkgs (clean free @ v0.0.1) + FEATURE_WIP
# pre-commit rejects external local replace; ./pkgs/shared keep-local allowed
root-linked freeHost ← clean external leaf
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next
  -> ready external pin first: wrk: cascade pin example.com/dot-pkgs @ v0.0.1
  -> then feature gen-commit (no external replace; hook OK)
  -> keep intra replace => ./pkgs/shared; exit 0
```

## Steps

1. Seed T-M1 fixture (`setupApplyPinBeforeFeatureMonorepoFreeHostExternal`):
   monorepo freeHost (intra owned-changed shared) + clean external replace +
   feature WIP + no-local-replace hook + modproxy.
2. Run apply with gen-commit + `--add-all` + `--merge-back` + `--tag-next`.
3. Assert ready external pin-before-feature, external replace dropped, intra kept.

## Context

- Encodes **monorepo freeHost** (same-label free pin dep) so pure multi-repo
  consumer deferral alone cannot cover this hole: freeHost peels early with
  replace still present → gen-commit hits hook (today RED).
- Ready external: clean free @ matching require `v0.0.1` (D3 keep-current).
- D7: separate pin auto-commit, then feature gen-commit.
- No `--push` (clean free; no residual cross-repo pending edges required).
- Do not rewrite sealed T1/T2 ASSERT meaning.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyPinBeforeFeatureMonorepoFreeHostExternal(t, req)
	// Flag set matches freeHost peel + gen-commit path; no --push.
	req.Args = cascadeUnwindGenCommitArgs(t, req,
		"--add-all",
		"--merge-back",
		"--tag-next",
	)
	return nil
}
```
