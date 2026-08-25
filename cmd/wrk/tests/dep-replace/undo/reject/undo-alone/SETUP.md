# Scenario

**Feature**: bare `--undo` is invalid without `--dep-replace`

```
wrk --undo
  -> non-zero
  -> --undo is only valid with --dep-replace
```

## Steps

1. Run `wrk --undo` alone.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--undo"}
	return nil
}
```
