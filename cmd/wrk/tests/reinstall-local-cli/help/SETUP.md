# Scenario

**Feature**: wrk root help documents --reinstall-local

```
wrk -h | wrk --help
  -> root usage mentions --reinstall-local
  -> exit 0
```

## Steps

- Descendants run root help (`-h` or `--help`).

```go
func Setup(t *testing.T, req *Request) error {
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```
