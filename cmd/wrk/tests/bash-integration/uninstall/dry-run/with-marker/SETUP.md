# Scenario

**Feature**: uninstall dry-run with markers previews removal without writing

```
pre-installed dual profile markers
wrk --bash-integration --uninstall --dry-run -> preview removal, profiles unchanged
```

## Steps

1. Pre-seed both profiles with wrk markers and unrelated content.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```