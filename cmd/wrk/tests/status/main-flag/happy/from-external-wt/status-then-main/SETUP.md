# Scenario

**Feature**: wrk --status --main from external wt (status flag first)

```
external wt cwd -> wrk --status --main -> same as --main --status
```

## Steps

1. Parent created main + external wt; cwd = external.
2. Args = `--status`, `--main` (order swapped).

```go
func Setup(t *testing.T, req *Request) error {
	setMainStatusArgs(req, "--status", "--main")
	return nil
}
```