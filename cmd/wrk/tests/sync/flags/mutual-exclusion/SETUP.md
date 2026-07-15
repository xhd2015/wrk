# Scenario

**Feature**: wrk --sync rejects non-composable mode flags

```
# --sync combined with --list / --status -> mutually exclusive error
# --done / --merge-back are composable (covered under monotree done-sync/ + merge-back-sync/)
wrk --sync --list|--status -> error before sync body
```

## Steps

- Descendants combine `--sync` with a non-composable mode flag (`--list`, `--status`).

```go
func Setup(t *testing.T, req *Request) error {
	syncEnsureHelpersUsed()
	return nil
}
```
