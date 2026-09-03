# Scenario

**Feature**: clean tree + `--gen-commit-msg --commit` with compose partner soft-skips

```
repo/ (clean)
  -> wrk --add-all --gen-commit-msg --commit --exec true
  -> exit 0
  -> stderr: notice: worktree clean, skip commit
  -> HEAD unchanged
```

## Steps

1. Init clean repo (seed commit only).
2. Run gen-commit compose with `--exec true` (no agent needed: empty index soft-skips before generate).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initCleanGitRepo(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{
		"--add-all",
		"--gen-commit-msg",
		"--commit",
		"--exec", "true",
	}
	return nil
}
```
