# Scenario

**Feature**: bare --all without --dep-update is rejected

```
wrk --all
  -> non-zero
  -> --all only valid with --dep-update (or unrecognized until flag lands)
```

## Steps

1. Run `wrk --all` alone (no `--dep-update`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--all"}
	return nil
}
```
