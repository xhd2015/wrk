# Scenario

**Feature**: `wrk --set-config -h` prints set-config dispatcher usage

```
workspace/ -> wrk --set-config -h
  -> dispatcher usage on stdout, exit 0
```

## Steps

1. Run `wrk --set-config -h` from neutral cwd (isolated WRK_HOME).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("-h")
	return nil
}
```
