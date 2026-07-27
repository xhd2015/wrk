# Scenario

**Feature**: `wrk --sync --dry-run` plans FF actions without mutating refs

```
# same plan as real sync but stdout lines prefixed "would: "
# no git merge; rev-parse before == after
main + linked wt -> wrk --sync --dry-run -> would: details + would: summary
```

## Steps

- Descendants build fixtures and pass `--dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
