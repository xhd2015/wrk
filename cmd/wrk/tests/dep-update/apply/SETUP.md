# Scenario

**Feature**: wrk --dep-update applies drop-replace + require@latest tag

```
consumer + tagged dep(s)
  -> wrk --dep-update <dir>…
  -> dep-update <mod> -> vN.N.N
  -> replace dropped; require set; no tidy
```

## Steps

- Leaves seed fixtures and set apply Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```
