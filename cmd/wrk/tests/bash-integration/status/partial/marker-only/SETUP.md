# Scenario

**Feature**: status reports partial when profile markers exist without bash.sh

```
dual profile markers present, bash.sh absent
wrk --bash-integration --status -> partial, exit 1
```

## Steps

1. Pre-seed both profiles with wrk markers only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```