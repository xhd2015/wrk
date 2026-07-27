# Scenario

**Feature**: wrk --repos runs from a git checkout or a subdirectory of one

```
git checkout cwd -> wrk --repos -> repo paths rooted at checkout top
```

## Steps

- Descendant scenarios create a git checkout and set `req.RepoDir` to the process cwd for invoking `wrk --repos`.

## Context

- `wrk --repos` resolves the git toplevel before repository discovery.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--repos"}
	return nil
}
```
