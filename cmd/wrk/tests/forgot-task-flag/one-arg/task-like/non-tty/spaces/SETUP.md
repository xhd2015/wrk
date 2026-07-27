# Scenario

**Feature**: one-arg multi-word, non-TTY from git repo cwd → error + hint

```
(cd myrepo && wrk "fix the login bug") non-TTY
  -> non-zero; task/source messaging; -t hint; no worktree
```

## Steps

1. Cwd = mainRepo.
2. TargetDir (first positional) = multi-word task text.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupOneArg(t, req, taskLikeSpaces)
	return nil
}
```
