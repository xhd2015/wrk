# Scenario

**Feature**: single -m value may contain newlines; first line is HEAD subject

```
repo/ (1 staged) -> wrk --commit -m "feat: subj\n\nbody line"
  -> exit 0
  -> HEAD subject = "feat: subj"
```

## Preconditions

- Locked decision D8: single `-m` value; newlines inside string OK (not multi `-m` flags).

## Steps

1. Stage `change.go`.
2. Run with a multi-line message string in one `-m` arg.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFile(t, req)
	msg := "feat: subj\n\nbody line"
	req.Args = []string{"--commit", "-m", msg}
	return nil
}
```
