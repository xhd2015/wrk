# Scenario

**Feature**: `wrk --bash-integration --install --help` prints usage and does not install

```
workspace/ -> wrk --bash-integration --install --help
  -> usage on stdout, exit 0; no filesystem writes
```

## Steps

1. Run `wrk --bash-integration --install --help` on a fresh isolated HOME/WRK_HOME.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.CLIArgs = []string{"--bash-integration", "--install", "--help"}
	return nil
}
```
