# Scenario

**Feature**: dir-mode dry-run plans skip tidy when vendor/ is present

```
nearest consumer has replace + require + empty vendor/
  -> wrk --dep-update <dep> --dry-run
  -> would: pin example.com/dep
  -> would: skip tidy  (vendor/)
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
