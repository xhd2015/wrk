# Scenario

**Feature**: wrk --gen-commit-msg --dry-run pure plan via library mock B

```
# dry-run host: --gen-commit-msg is on the wrk --dry-run allow-list
git repo with staged files
  -> wrk --gen-commit-msg --dry-run
  -> stdout mock B for N staged files; exit 0; no agent
  -> --model accepted and unused
  -> binaries: would-unstage on stderr; index unchanged
  -> --commit: would: git commit on stderr; HEAD unchanged
  -> --commit --no-verify: would-line includes --no-verify
```

## Preconditions

- Root harness has initialized WorkRoot / WrkHome.
- Leaves call `stageOneTextFile` / `stageBinaryAndTextFile` (or equivalent) before setting Args.

## Steps

1. Inherit root Setup.
2. Leaf stages files and sets `--gen-commit-msg --dry-run` (+ optional flags).

## Context

- Mock B: `dry-run: would generate commit message for N staged file(s)\n`
- N = staged count **before** unstage (N=1 for text-only leaves; N=2 for binary+text).
- Would-unstage / would-commit plan lines go to stderr only.
- Process cwd is the git repo (`req.RepoDir`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: dry-run leaves share stage + mock B helpers.
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
