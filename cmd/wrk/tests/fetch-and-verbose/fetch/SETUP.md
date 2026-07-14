# Scenario

**Feature**: wrk --fetch flag validation and scoped upstream refresh

```
# bare --fetch or with invalid modes -> stderr error, exit 1
wrk --fetch [--list|--done] -> error

# --fetch with --projects or --status refreshes upstream before Remote: comparison
wrk --projects [--fetch] -> Remote: from local refs or post-fetch
```

## Steps

- Descendants set `req.Args` for `--fetch` combined with `--projects` or `--status`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureFetchVerboseHelpersUsed()
	return nil
}
```