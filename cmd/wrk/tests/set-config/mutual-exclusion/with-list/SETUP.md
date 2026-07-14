# Scenario

**Feature**: set-config cannot combine with `--list`

```
wrk --set-config --create --new-terminal --list -> non-zero
```

## Steps

1. Run conflicting flags.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = setConfigArgs("--create", "--new-terminal", "--list")
	return nil
}
```
