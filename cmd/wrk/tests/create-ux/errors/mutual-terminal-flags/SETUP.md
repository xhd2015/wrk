# Scenario

**Feature**: terminal mode flags are mutually exclusive

```
wrk --new-terminal --reuse-terminal -> non-zero
```

## Steps

1. Run two terminal mode flags together.

```go
func Setup(t *testing.T, req *Request) error {
	installCreateUXMocks(t, req, "darwin")
	req.Args = []string{"--new-terminal", "--reuse-terminal"}
	return nil
}
```
