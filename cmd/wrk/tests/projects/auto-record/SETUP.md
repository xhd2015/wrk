# Scenario

**Feature**: auto-record main repo on every wrk invocation

```
# effective work dir resolves to git main repo -> record in projects.json (source: auto)
wrk [dir] [mode] -> auto-record main repo path

# no record when work dir missing or not git
non-git cwd / missing <dir> -> no projects.json entry
```

## Steps

- Descendants vary cwd vs `<dir>` arg and whether the work dir is valid git.
- Most scenarios use `wrk --list` to trigger auto-record without worktree side effects.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureProjectsHelpersUsed()
	req.Args = []string{"--list"}
	return nil
}
```