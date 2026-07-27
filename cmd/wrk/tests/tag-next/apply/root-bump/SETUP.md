# Scenario

**Feature**: apply creates lightweight v0.0.2 tag at HEAD

```
# v0.0.1 tagged, README changed -> wrk --tag-next -> git tag v0.0.2 HEAD
git repo + tags -> wrk --tag-next -> tagged line + 1 tag created
```

## Steps

1. `setupRootBumpRepo`.
2. Run `wrk --tag-next`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupRootBumpRepo(t, req)
	req.Args = []string{"--tag-next"}
	return nil
}
```