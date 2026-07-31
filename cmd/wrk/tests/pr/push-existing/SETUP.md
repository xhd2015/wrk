# Scenario

**Feature**: `wrk --pr --push` **push-existing** — open PR required → full tip push → URL; optional `--comment` after push; never create

```
# push-existing (P3 classic TDD — RED until implementer)
linked wt + github origin + gh; --push + --pr; no --title
  -> wrk --pr --push  [optional --comment C]
  -> list open PR for head FIRST (before any push)
  -> open PR: full tip push (origin advances when local ahead) + URL
  -> with --comment: then gh pr comment; stdout pushed + comment added + URL
  -> no open PR: non-zero; stderr no open pull request; origin tip UNCHANGED
  -> never gh pr create; no title-ignored warning

# flag order free
  --push --pr same as --pr --push
```

## Preconditions

- Inherits `pr/SETUP.md` fake `gh`, github-shaped origin + bare pushurl, `setupPrLinkedFeatureRemoteExistsLocalAhead` (local ahead + `origin-feature-before` snapshot).
- Classic TDD P3: product still rejects bare `--pr --push` without title (“title is required”). Leaves **RED** until implementer.
- Create compose with title stays under `compose/push-pr/` (unchanged GREEN).
- L2 `InProcess = true`; parallel-safe (`PathPrepend` / `ExtraEnv` only).

## Steps

- Leaves seed local-ahead fixture, install fake gh, set open-PR list or leave empty, set `req.Args`.

## Context

- Full tip push uses the same confirm line as compose `--push`: `pushed <branch> → origin/<branch>`.
- Multi-stage stdout may insert a blank line between stages; asserts use `joinStdoutBlocks`.
- “List before push” is load-bearing: no open PR must leave origin tip equal to pre-run snapshot.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	prPushExistingEnsureHelpers()
	return nil
}

// prPushExistingArgs is bare push-existing: --pr --push, no title/comment.
func prPushExistingArgs() []string {
	return []string{"--pr", "--push"}
}

// prPushThenPrArgs: argv order free — --push before --pr (execution still list→push→…).
func prPushThenPrArgs() []string {
	return []string{"--push", "--pr"}
}

// prPushExistingCommentArgs is push then comment: --pr --push --comment C (no title).
func prPushExistingCommentArgs(comment string) []string {
	if comment == "" {
		comment = prDefaultComment
	}
	return []string{"--pr", "--push", "--comment", comment}
}

// prPushThenPrCommentArgs: --push --pr --comment C (flag order free).
func prPushThenPrCommentArgs(comment string) []string {
	if comment == "" {
		comment = prDefaultComment
	}
	return []string{"--push", "--pr", "--comment", comment}
}

// prPushExistingStdout is multi-stage: full-push confirm, blank line, URL only.
func prPushExistingStdout(branch, url string) string {
	return joinStdoutBlocks(prPushConfirmLine(branch), prShowStdout(url))
}

// prPushExistingCommentStdout is full-push confirm, blank line, comment-added + URL.
func prPushExistingCommentStdout(branch, url string) string {
	return joinStdoutBlocks(prPushConfirmLine(branch), prExistingStdout(url))
}

func prPushExistingEnsureHelpers() {
	_ = prPushExistingArgs
	_ = prPushThenPrArgs
	_ = prPushExistingCommentArgs
	_ = prPushThenPrCommentArgs
	_ = prPushExistingStdout
	_ = prPushExistingCommentStdout
}
```
