# Scenario

**Feature**: wrk --dep-replace hard-fails on missing dir, non-module dep, or no consumer go.mod

```
invalid dep path | plain dir without go.mod | workDir with no go.mod uptree
  -> wrk --dep-replace …
  -> Error non-zero; no successful replace line
```

## Steps

- Leaves seed minimal fixtures and invalid/edge dep paths.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepReplaceHelpersUsed()
	return nil
}
```
