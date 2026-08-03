# Scenario

**Feature**: clean tree + --commit -m fails with nothing to commit

```
repo/ (clean) -> wrk --commit -m "feat: nothing"
  -> non-zero
  -> nothing to commit / no staged
```

## Preconditions

- Isolated git repo with only the seed commit; nothing staged; no untracked needed.

## Steps

1. Init clean repo via `initCleanGitRepo`.
2. Run `wrk --commit -m "feat: nothing"`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initCleanGitRepo(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--commit", "-m", "feat: nothing"}
	return nil
}
```
