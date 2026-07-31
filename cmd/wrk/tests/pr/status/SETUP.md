# Scenario

**Feature**: `wrk --pr --status` **PR status** — open PR metadata + checks/reviews rollup; read-only

```
# status (P4 classic TDD — RED until implementer)
linked wt + github origin + gh; --pr --status (no --title / --comment / --push)
  -> shared gates (linked wt, github.com origin, named branch, gh on PATH)
  -> list open PR for head (open only)
  -> open PR: gh pr view N --json number,title,url,state,isDraft,reviewDecision,statusCheckRollup
  -> stdout:
       URL
       State:     open|draft|…
       Title:     …
       Checks:    success|failure|pending|none|mixed
       Reviews:   approved|changes requested|review required|none
  -> exit 0 even when Checks=failure (report, not gate)
  -> no open PR: exit 0; stderr warning: no open …; stdout empty
  -> never ensure-push / pr create / pr comment

# invalid combinations
  --pr --status + --title/--comment → non-zero
  --pr --status + --push → non-zero
  flag order free: --status --pr same as --pr --status
```

## Preconditions

- Inherits `pr/SETUP.md` fake `gh` (list + **view** via `FAKE_GH_VIEW_JSON`), github-shaped origin + bare pushurl, linked feature fixtures, and helpers: `prStatusArgs`, `prStatusThenPrArgs`, `prStatusStdout`, `setFakeGhViewJSON`, `prViewJSON`, `prRollup*JSON`.
- Classic TDD P4: product still treats `status && prFlag` as mutual exclusion (no `runPRStatus`). Leaves **RED** until implementer carves out the compose and implements status output.
- Shared gates identical to show/comment/push-existing (linked worktree, github.com origin, named branch, `gh` on PATH).
- L2 `InProcess = true`; parallel-safe (`PathPrepend` / `ExtraEnv` only).
- Global bare `wrk --status` (git worktree status) is **not** this tree; must remain GREEN elsewhere.

## Steps

- Open-PR leaves: `setupPrLinkedFeatureRemoteExists` + `installFakeGh` + `setFakeGhExistingPR` + `setFakeGhViewJSON(prViewJSON(…))` + `req.Args = prStatusArgs()` (or reverse order).
- No-open leaf: same fixture, default list `[]`, no view JSON override required.
- Refuse leaves: linked fixture + invalid argv; fake gh optional (should fail before gh).

## Context

- **Checks** rollup from `statusCheckRollup` only (not full `gh pr checks` table). Fixture fragments: `prRollupSuccessJSON` / `Failure` / `Pending` / `None`.
- **Reviews** map `reviewDecision` → human lower case (`REVIEW_REQUIRED` → `review required`, empty → `none`).
- **State**: `isDraft` → `draft`; else `state` lowercased (`OPEN` → `open`).
- Field labels and spacing pinned by `prStatusStdout` (column-aligned values); flexible whitespace OK if product matches labels/values.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
