# Scenario

**Feature**: wrk root help documents --dep-replace and unwind/stack

```
wrk -h
  -> root usage mentions --dep-replace
  -> root usage mentions unwind/stack for --dep-replace
```

## Steps

- Descendants run help and assert flag documentation.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepReplaceHelpersUsed()
	return nil
}
```
