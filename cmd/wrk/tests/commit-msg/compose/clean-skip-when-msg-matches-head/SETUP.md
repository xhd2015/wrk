# Scenario

**Feature**: clean tree + compose `--commit -m` soft-skips when message already matches HEAD

```
repo/ (clean, HEAD subject "initial")
  -> wrk --commit -m "initial" --exec true
  -> exit 0
  -> stderr: notice: worktree clean, skip commit
  -> HEAD unchanged; exec runs
```

## Steps

1. Init clean repo (seed commit message `initial`).
2. Run composed manual commit with `-m` equal to HEAD message + `--exec true`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initCleanGitRepo(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--commit", "-m", "initial", "--exec", "true"}
	return nil
}
```
