# Scenario

**Feature**: `--new-window` alone under set-config also persists terminal.mode=new

```
wrk --set-config --create --new-window
  -> window.mode=new AND terminal.mode=new
```

## Steps

1. No prior config.
2. Run `--set-config --create --new-window` only.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--create", "--new-window")
	return nil
}
```
