# Scenario

**Feature**: wrk --pin-locals --dry-run prints would: lines without writes

```
stack fixture -> wrk --pin-locals --dry-run
  -> would: pin-local … (when work pending)
  -> go.mod unchanged
  -> no go mod tidy
  -> exit 0
```

## Steps

- Descendants seed stack fixtures and set Args to `--pin-locals --dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals", "--dry-run"}
	ensurePinLocalsHelpersUsed()
	return nil
}
```
