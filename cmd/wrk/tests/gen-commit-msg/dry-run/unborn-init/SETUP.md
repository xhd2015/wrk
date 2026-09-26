# Scenario

**Feature**: `wrk --gen-commit-msg --dry-run --commit` on unborn HEAD plans commit

```
# git init only; one staged file; exclusive-branch guard must not fatal
unborn repo (1 staged) -> wrk --gen-commit-msg --dry-run --commit
  -> exit 0
  -> would: git commit
  -> HEAD stays unborn (no commit)
```

## Preconditions

- Isolated git repo after `git init` (no commits); one staged text file.
- Hooks disabled.

## Steps

1. Init unborn repo and stage `change.go`.
2. Run `wrk --gen-commit-msg --dry-run --commit`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFileUnborn(t, req)
	req.Args = []string{"--gen-commit-msg", "--dry-run", "--commit"}
	return nil
}
```
