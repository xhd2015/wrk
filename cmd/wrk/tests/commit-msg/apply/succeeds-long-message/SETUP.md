# Scenario

**Feature**: --message long form commits with the supplied subject

```
repo/ (1 staged) -> wrk --commit --message "feat: long form"
  -> exit 0
  -> HEAD subject = "feat: long form"
```

## Preconditions

- Same apply path as short `-m`; long flag form only.

## Steps

1. Stage `change.go`.
2. Run `wrk --commit --message "feat: long form"`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFile(t, req)
	req.Args = []string{"--commit", "--message", "feat: long form"}
	return nil
}
```
