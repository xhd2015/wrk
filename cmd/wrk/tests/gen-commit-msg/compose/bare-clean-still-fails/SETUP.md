# Scenario

**Feature**: bare `--gen-commit-msg --commit` on clean tree still hard-fails

```
repo/ (clean)
  -> wrk --add-all --gen-commit-msg --commit
  -> non-zero
  -> no staged / nothing to commit
  -> no soft-skip notice
```

## Steps

1. Init clean repo.
2. Run bare gen-commit with `--commit` (no compose partner).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initCleanGitRepo(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--add-all", "--gen-commit-msg", "--commit"}
	return nil
}
```
