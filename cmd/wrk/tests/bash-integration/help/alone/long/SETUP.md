# Scenario

**Feature**: `wrk --bash-integration --help` prints dedicated bash-integration usage

```
workspace/ -> wrk --bash-integration --help
  -> usage on stdout, exit 0
```

## Steps

1. Run `wrk --bash-integration --help` from neutral cwd (isolated WRK_HOME).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.CLIArgs = []string{"--bash-integration", "--help"}
	return nil
}
```
