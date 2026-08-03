# Scenario

**Feature**: wrk-owned manual commit apply path with -m/--message

```
# staged + message → real git commit (HEAD subject)
# --dry-run → would-lines only
# --no-verify → skip failing hooks
# no staged → error
# --add-all → stage then commit
```

## Preconditions

- Root harness WorkRoot / WrkHome.
- Leaves init isolated git repos (hooks disabled unless no-verify leaf).

## Steps

1. Inherit root Setup.
2. Leaf stages/places files, sets Args with `--commit` + message.

## Context

- Prefer InProcess Capture with RepoDir = git repo.
- Assert HEAD subject via `gitHEADSubject` after successful commit.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureCommitMsgHelpersUsed()
	return nil
}
```
