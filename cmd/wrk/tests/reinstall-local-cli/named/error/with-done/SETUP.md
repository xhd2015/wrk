# Scenario

**Feature**: names are rejected when composed with `--done`

```
wrk --done --reinstall-local tool -> non-zero exclusive-names error
```

## Steps

1. No module fixtures required (validation before plan).
2. Run `--done --reinstall-local tool`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--done", "--reinstall-local", "tool"}
	return nil
}
```
