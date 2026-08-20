# Scenario

**Feature**: dry-run identifies the overridden Go SDK without running tidy

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVersionedTidy(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```
