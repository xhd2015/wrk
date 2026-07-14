# Scenario

**Feature**: install dry-run when fully installed reports is up to date

```
pre-install real assets (current script + dual profile markers)
wrk --bash-integration --install --dry-run -> is up to date, no changes
```

## Steps

1. Pre-seed profile markers with unrelated content.
2. Pre-install so bash.sh matches the embedded script.
3. Run install dry-run.

```go
func Setup(t *testing.T, req *Request) error {
// Markers + unrelated profile content; PreInstall writes current bash.sh
// and ensures markers exist (append is a no-op if already present).
req.PreExistingBashProfile = preInstalledProfileContent()
req.PreExistingBashRC = preInstalledProfileContent()
req.PreInstall = true
return nil
}
```
