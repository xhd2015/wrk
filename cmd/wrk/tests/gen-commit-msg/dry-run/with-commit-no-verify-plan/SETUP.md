# Scenario

**Feature**: wrk --gen-commit-msg --dry-run --commit --no-verify plans commit with --no-verify

```
# dry-run commit plan includes --no-verify through wrk binary
staged -> wrk --gen-commit-msg --dry-run --commit --no-verify
  -> stderr would-line includes --no-verify
  -> HEAD unchanged
```

## Preconditions

- One staged text change.
- HEAD subject recorded before the run.

## Steps

1. Stage one text file.
2. Record pre-run HEAD subject.
3. Run `wrk --gen-commit-msg --dry-run --commit --no-verify`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFile(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--gen-commit-msg", "--dry-run", "--commit", "--no-verify"}
	return nil
}
```
