# Scenario

**Feature**: wrk --status runs from a git checkout or a subdirectory of one

```
# cwd inside git checkout resolves to its checkout root
git checkout cwd -> wrk --status -> status blocks rooted at checkout top
```

## Preconditions

- The effective cwd must be inside a git work tree.

## Steps

- Descendant scenarios create one or more git repositories under the fixture root.
- `req.RepoDir` is the process cwd for invoking `wrk --status`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--status"}
	return nil
}
```
