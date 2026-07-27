# Scenario

**Feature**: auto-record when effective work dir comes from `<dir>` positional

```
# shell cwd is WorkRoot; <dir> selects the git checkout
WorkRoot -> wrk <dir> --list -> projects.json records <dir>'s main repo
```

## Steps

- Set `req.TargetDir` to the target path and `req.RepoDir` to `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```