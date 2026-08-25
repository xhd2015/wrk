# Scenario

**Feature**: `wrk --bash-integration -h` prints dedicated bash-integration usage

```
workspace/ -> wrk --bash-integration -h
  -> usage on stdout, exit 0
```

## Steps

1. Run `wrk --bash-integration -h` from neutral cwd (isolated WRK_HOME).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.CLIArgs = []string{"--bash-integration", "-h"}
	return nil
}
```
