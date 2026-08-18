# Scenario

**Feature**: --all pins an inventory-owned outdated require on another stack checkout

```
# primary + external/kool both require example.com/lib@v1.0.0
# lib registered at v1.2.3
cwd=primary -> wrk --dep-update --all
  -> no argv dep headers
  -> pin + tidy app and kool
```

## Steps

1. Seed stack + registered owner + file:// GOPROXY.
2. Run `--all` apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllStackOutdated(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```
