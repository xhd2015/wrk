# Scenario

**Feature**: wrk root help documents --version

```
wrk -h | wrk --help
  -> root usage mentions --version
  -> exit 0
```

## Steps

- Descendants run root help (`-h` or `--help`).

```go
func Setup(t *testing.T, req *Request) error {
	ensureVersionHelpersUsed()
	return nil
}
```