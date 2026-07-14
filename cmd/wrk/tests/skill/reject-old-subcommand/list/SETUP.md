# Scenario

**Feature**: wrk skill list (old subcommand) is rejected

```
workspace/ -> wrk skill list -> non-zero; unknown / expected flag action
```

## Steps

1. Run `wrk skill list` (positional subcommand, not `--list`) from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "list"}
	return nil
}
```
