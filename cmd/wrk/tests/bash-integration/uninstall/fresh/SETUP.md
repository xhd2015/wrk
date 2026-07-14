# Scenario

**Feature**: uninstall removes dual profile markers and preserves bash.sh

```
pre-installed bash.sh + dual profile markers
wrk --bash-integration --uninstall -> markers gone, bash.sh intact
```

## Steps

1. Pre-seed installed state in both profiles and bash.sh.
2. Run uninstall.

```go
func Setup(t *testing.T, req *Request) error {
	req.PreExistingBashSh = preInstalledBashShContent()
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```