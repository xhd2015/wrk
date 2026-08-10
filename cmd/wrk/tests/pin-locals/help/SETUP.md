# Scenario

**Feature**: wrk root help documents --pin-locals

```
wrk -h | wrk --help
  -> root usage mentions --pin-locals
  -> exit 0
```

## Steps

- Descendants run root help (`-h` or `--help`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensurePinLocalsHelpersUsed()
	return nil
}
```
