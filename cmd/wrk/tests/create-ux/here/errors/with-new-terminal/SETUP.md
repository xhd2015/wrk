# Scenario

**Feature**: `--here --new-terminal` is mutually exclusive

```
wrk --here --new-terminal -> non-zero
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	installCreateUXMocks(t, req, "darwin")
	req.Args = []string{"--here", "--new-terminal"}
	return nil
}
```
