# Scenario

**Feature**: wrk --dep-update applies drop-replace + require@latest + versioned tidy

```
consumer + tagged dep(s)
  -> wrk --dep-update <dir>…
  -> ==== dep-update ====; dep headers; checkout → module → pin
  -> replace dropped; require set
  -> go mod tidy ok | skip tidy  (vendor/)
  -> dep-update: updated N modules in C checkouts
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
