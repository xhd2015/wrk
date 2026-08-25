# Scenario

**Feature**: help with `--install` still prints usage and never mutates

```
wrk --bash-integration --install -h|--help
wrk --bash-integration --help --install
  -> usage on stdout, exit 0
  -> no bash.sh / profile marker writes
```

## Steps

- Leaves set action + help order under test.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireNoPreseed(t, req)
	return nil
}
```
