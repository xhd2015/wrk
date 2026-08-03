# Scenario

**Feature**: --message alone requires --commit

```
workspace/ -> wrk --message "feat: alone"
  -> non-zero; message about requires --commit
```

## Preconditions

- Long form `--message` has the same host requirement as short `-m`.

## Steps

1. Run `wrk --message "feat: alone"` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--message", "feat: alone"}
	return nil
}
```
