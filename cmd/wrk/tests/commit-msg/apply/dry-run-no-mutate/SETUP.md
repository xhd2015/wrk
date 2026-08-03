# Scenario

**Feature**: --dry-run with manual message plans git commit without mutating HEAD

```
staged -> wrk --commit -m "feat: x" --dry-run
  -> stderr: would: git commit (and message)
  -> HEAD subject unchanged
```

## Preconditions

- One staged text change; HEAD subject is the initial commit subject.

## Steps

1. Stage one text file.
2. Record pre-run HEAD subject.
3. Run `wrk --commit -m "feat: x" --dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFile(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--commit", "-m", "feat: x", "--dry-run"}
	return nil
}
```
