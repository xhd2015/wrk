# Scenario

**Feature**: `wrk --set-config --help` prints set-config dispatcher usage

```
workspace/ -> wrk --set-config --help
  -> dispatcher usage on stdout, exit 0
```

## Steps

1. Run `wrk --set-config --help` from neutral cwd (isolated WRK_HOME).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--help")
	return nil
}
```
