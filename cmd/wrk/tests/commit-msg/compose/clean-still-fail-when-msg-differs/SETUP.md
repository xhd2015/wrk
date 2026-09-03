# Scenario

**Feature**: clean tree + compose `--commit -m` still fails when message differs from HEAD

```
repo/ (clean, HEAD "initial")
  -> wrk --commit -m "feat: other" --exec true
  -> non-zero
  -> no staged / nothing to commit
  -> no soft-skip notice
```

## Steps

1. Init clean repo.
2. Run composed manual commit with a message that is not HEAD's message.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initCleanGitRepo(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--commit", "-m", "feat: other", "--exec", "true"}
	return nil
}
```
