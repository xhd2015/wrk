# Scenario

**Feature**: auto-record when effective work dir is process cwd (no `<dir>` arg)

```
# process cwd inside git checkout -> record resolved main repo
cwd (main or subdir) -> wrk --list -> projects.json
```

## Steps

- Descendants set `req.RepoDir` to main repo root or a nested subpath.

```go
func Setup(t *testing.T, req *Request) error {
	ensureProjectsHelpersUsed()
	return nil
}
```