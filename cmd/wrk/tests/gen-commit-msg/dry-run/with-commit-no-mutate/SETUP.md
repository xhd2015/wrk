# Scenario

**Feature**: wrk --gen-commit-msg --dry-run --commit plans git commit without mutating HEAD

```
# dry-run + commit flag: would-line only through wrk binary
staged -> wrk --gen-commit-msg --dry-run --commit
  -> stderr: would: git commit -m '…'
  -> HEAD subject unchanged
```

## Preconditions

- One staged text change; HEAD subject is the initial commit subject.

## Steps

1. Stage one text file in an isolated repo.
2. Record pre-run HEAD subject.
3. Run `wrk --gen-commit-msg --dry-run --commit`.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--gen-commit-msg", "--dry-run", "--commit"}
	return nil
}
```
