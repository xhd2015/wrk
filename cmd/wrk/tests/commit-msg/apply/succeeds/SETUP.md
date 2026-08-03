# Scenario

**Feature**: staged changes + --commit -m "feat: x" creates commit with that subject

```
repo/ (1 staged) -> wrk --commit -m "feat: x"
  -> exit 0
  -> HEAD subject = "feat: x"
```

## Preconditions

- Isolated git repo with hooks disabled; one staged text file.

## Steps

1. Stage `change.go` via `stageOneTextFile`.
2. Run `wrk --commit -m "feat: x"`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFile(t, req)
	req.Args = []string{"--commit", "-m", "feat: x"}
	return nil
}
```
