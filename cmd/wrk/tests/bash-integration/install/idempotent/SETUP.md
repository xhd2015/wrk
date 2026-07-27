# Scenario

**Feature**: second install is a no-op

```
pre-seeded bash.sh + dual profile markers
wrk --bash-integration --install (twice) -> single marker per profile, script preserved
```

## Steps

1. Pre-seed bash.sh and both profile markers.
2. Run install twice (`req.RunTwice = true`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PreExistingBashSh = preInstalledBashShContent()
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	req.RunTwice = true
	return nil
}
```