# Scenario

**Feature**: status reports partial when profile markers exist without bash.sh

```
dual profile markers present, bash.sh absent
wrk --bash-integration --status -> partial, exit 1
```

## Steps

1. Pre-seed both profiles with wrk markers only.

```go
func Setup(t *testing.T, req *Request) error {
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```