# Scenario

**Feature**: `wrk --pr --comment C` **comment-only** — attach additive comment to existing open PR; never create; never push

```
# comment-only mode (P2 classic TDD — RED until implementer)
linked wt + github origin + gh on PATH; --comment set; no --title
  -> wrk --pr --comment C
  -> gh pr list --head <branch> --state open
  -> open PR: gh pr comment <n> --body C; stdout comment added + URL
  -> no open PR: non-zero; stderr mentions no open pull request
  -> never ensure-push / gh pr create; no title-ignored warning

# empty --comment after trim
  -> non-zero before gh comment/create
```

## Steps

- Leaves seed fixtures, install fake gh, set `req.Args = prCommentOnlyArgs(...)`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
