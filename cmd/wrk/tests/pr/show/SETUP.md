# Scenario

**Feature**: bare `wrk --pr` **show** mode — print open PR URL for head branch, or empty stdout

```
# linked wt + github origin + gh on PATH; no --title / --comment / --status
linked wt (feature) + fake gh list
  -> wrk --pr
  -> gh pr list --head <branch> --state open
  -> open PR: stdout = URL\n only (exit 0)
  -> no open PR: empty stdout (exit 0)
  -> never ensure-push / pr create / pr comment

# same hard refuse gates as create (main repo / non-github / …)
  -> non-zero; no create/comment
```

## Steps

- Leaves seed fixtures, install fake gh, set `req.Args = prShowArgs()` (or refuse fixtures).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
