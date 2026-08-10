# Scenario

**Feature**: wrk root help documents --dep-update

```
wrk -h
  -> root usage mentions --dep-update
```

## Steps

- Descendants run help and assert flag documentation.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```
