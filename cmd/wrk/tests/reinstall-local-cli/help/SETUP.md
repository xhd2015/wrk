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
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```
