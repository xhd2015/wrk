# Scenario

**Feature**: `wrk --set-config --show -h` prints show-level usage

```
workspace/ -> wrk --set-config --show -h
  -> show usage on stdout, exit 0
```

## Steps

1. Run `wrk --set-config --show -h`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--show", "-h")
	return nil
}
```
