# Scenario

**Feature**: dir-mode dry-run plans skip tidy when vendor/ is present

```
nearest consumer has replace + require + empty vendor/
  -> wrk --dep-update <dep> --dry-run
  -> would: dep-update example.com/dep -> v0.0.2
  -> would: skip tidy  module example.com/consumer  (vendor/)
  -> go.mod unchanged; vendor/ untouched
```

## Steps

1. Seed drop-replace-latest + empty `vendor/`.
2. Run with `--dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVendorSkipDir(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```
