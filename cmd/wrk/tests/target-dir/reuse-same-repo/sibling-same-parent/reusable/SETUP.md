# Scenario

**Feature**: same-parent sibling that is reusable (clean + same HEAD as source)

```
# reusable sibling under {WorkRoot}/target
# TTY -> would reuse / skip creating prompt; non-TTY -> create (no refuse)
myrepo + clean same-HEAD sibling under same parent
  -> wrk myrepo {WorkRoot}/target
```

## Steps

- Descendants add reusable sibling(s); split TTY vs non-TTY.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureNamedBringReuseHelpersUsed()
	return nil
}
```
