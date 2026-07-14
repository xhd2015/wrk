# Scenario

**Feature**: `--new-window --no-new-terminal` is an error

```
wrk --new-window --no-new-terminal -> non-zero (window requires terminal open path)
```

## Steps

1. Install mocks; run conflicting flags.

```go
func Setup(t *testing.T, req *Request) error {
	installCreateUXMocks(t, req, "darwin")
	req.Args = []string{"--new-window", "--no-new-terminal"}
	return nil
}
```
