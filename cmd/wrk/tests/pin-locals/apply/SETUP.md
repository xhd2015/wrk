# Scenario

**Feature**: wrk --pin-locals applies relative replaces then go mod tidy

```
stack fixture -> wrk --pin-locals
  -> pin-local … lines for each add/rewrite
  -> relative replace in go.mod
  -> go mod tidy per edited consumer
  -> summary: pin-locals: applied N, tidy ok M, tidy failed F
  -> tidy soft-fail: warning: + continue + exit 0
```

## Steps

- Descendants seed fixtures and set Args to `--pin-locals`.
- Offline env (`GOPROXY=off`) applied by setup helpers.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--pin-locals"}
	ensurePinLocalsHelpersUsed()
	return nil
}
```
