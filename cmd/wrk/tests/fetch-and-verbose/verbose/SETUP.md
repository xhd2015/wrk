# Scenario

**Feature**: wrk -v/--verbose logs major git subprocesses to stderr only

```
-v with any mode -> stderr [timestamp] $ git <major-subcommand>
minor reads (rev-parse, status, worktree list) -> no log
stdout content unchanged vs non-verbose run
```

## Steps

- Descendants set `req.Args` with `-v` or `--verbose` as required.

```go
func Setup(t *testing.T, req *Request) error {
	ensureFetchVerboseHelpersUsed()
	return nil
}
```