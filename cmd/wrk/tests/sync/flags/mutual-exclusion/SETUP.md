# Scenario

**Feature**: wrk --sync rejects other mode flags

```
# --sync combined with --done / --list / --status -> mutually exclusive error
wrk --sync --done|--list|--status -> error before sync body
```

## Steps

- Descendants combine `--sync` with another standalone mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	syncEnsureHelpersUsed()
	return nil
}
```
