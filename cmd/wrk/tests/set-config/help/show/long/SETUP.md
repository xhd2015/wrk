# Scenario

**Feature**: `wrk --set-config --show --help` prints show-level usage

```
workspace/ -> wrk --set-config --show --help
  -> show usage on stdout, exit 0
```

## Steps

1. Run `wrk --set-config --show --help` with empty config (help only).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--show", "--help")
	return nil
}
```
