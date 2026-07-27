# Scenario

**Feature**: only sub/ scope bumps when child-owned files changed

```
# root v0.0.1 + sub/v0.2.3 at baseline; only sub/lib.go changed -> sub/v0.2.4
git repo + scoped tags -> wrk --tag-next -> root skip, sub bump
```

## Steps

1. `setupSubScopeOnlyRepo`.
2. Run `wrk --tag-next`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSubScopeOnlyRepo(t, req)
	req.Args = []string{"--tag-next"}
	return nil
}
```