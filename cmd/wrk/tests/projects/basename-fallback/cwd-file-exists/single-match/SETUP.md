# Scenario

**Feature**: file in cwd plus one saved project yields concrete-path guided hint

```
workspace/myrepo (file) + saved/myrepo in projects.json
wrk myrepo -t 'desc' -> guided stderr with saved path hint; no worktree
```

## Steps

- Descendants seed exactly one saved project whose basename matches the cwd file name.
- Run create-mode `wrk <basename>` with flags preserved in the hint.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureCwdFileExistsHelpersUsed()
	return nil
}
```