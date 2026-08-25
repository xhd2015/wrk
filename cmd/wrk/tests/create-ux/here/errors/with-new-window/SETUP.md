# Scenario

**Feature**: `--here --new-window` is mutually exclusive

```
wrk --here --new-window -> non-zero
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	installCreateUXMocks(t, req, "darwin")
	req.Args = []string{"--here", "--new-window"}
	return nil
}
```
