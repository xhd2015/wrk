# Scenario

**Feature**: `--push` + force + `--dry-run` plans force-with-lease without mutating origin

```
myrepo (main) + origin
  -> wrk --push -f --dry-run  (or --force / flag-order variants)
  -> would: git push --force-with-lease origin main
  -> origin/main unchanged; no "pushed …" confirm
```

## Steps

- Grouping: leaves seed main+origin, snapshot origin tip, set force+push dry-run Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.InProcess = true
	return nil
}
```
