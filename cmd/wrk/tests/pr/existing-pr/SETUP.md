# Scenario

**Feature**: when an open PR already exists for the head branch, title is ignored and comment is still added

```
# open PR exists for feature-pr
linked wt + github origin + fake gh (list returns PR)
  -> wrk --pr --title NEW --comment C
  -> stderr warning: title ignored (PR already exists); existing title: …
  -> no gh pr create; gh pr comment still runs
  -> stdout: comment added + URL (no PR created / title set)
```

## Steps

- Leaves install fake gh with canned list JSON and set argv.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
