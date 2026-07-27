# Scenario

**Feature**: `--new-window --no-new-terminal` is an error

```
wrk --new-window --no-new-terminal -> non-zero (window requires terminal open path)
```

## Steps

1. Install mocks; run conflicting flags.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	installCreateUXMocks(t, req, "darwin")
	req.Args = []string{"--new-window", "--no-new-terminal"}
	return nil
}
```
