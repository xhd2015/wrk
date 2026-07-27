# Scenario

**Feature**: set-config cannot combine with `--list`

```
wrk --set-config --create --new-terminal --list -> non-zero
```

## Steps

1. Run conflicting flags.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = setConfigArgs("--create", "--new-terminal", "--list")
	return nil
}
```
