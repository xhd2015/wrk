# Scenario

**Feature**: same-parent sibling that is not reusable → create, no Policy B banner

```
# dirty OR clean-but-differs-from-source sibling under same parent
# -> create as today; no would-reuse / skip-creating
myrepo + non-reusable sibling under {WorkRoot}/target
  -> wrk myrepo {WorkRoot}/target -> new named subdir
```

## Steps

- Leaves mark sibling dirty or commit-ahead; leave non-TTY (default pipe).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	return nil
}
```
