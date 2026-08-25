# Scenario

**Feature**: root `wrk --help` points at bash-integration usage

```
workspace/ -> wrk --help
  -> root usage mentions Bash integration / --bash-integration
  -> points at wrk --bash-integration --help
```

## Steps

1. Run `wrk --help` from neutral cwd (isolated WRK_HOME).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.CLIArgs = []string{"--help"}
	return nil
}
```
