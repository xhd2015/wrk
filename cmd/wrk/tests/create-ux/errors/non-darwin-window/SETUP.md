# Scenario

**Feature**: window on non-darwin surfaces a clear platform error

```
DOT_PKGS_SPACE_GOOS=linux; wrk --new-window
  -> non-zero; stderr mentions macOS/darwin/unsupported
```

## Steps

1. Install mocks with goos=linux.
2. Run `--new-window`.

```go
func Setup(t *testing.T, req *Request) error {
	installCreateUXMocks(t, req, "linux")
	req.Args = []string{"--new-window"}
	return nil
}
```
