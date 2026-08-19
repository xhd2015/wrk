# Scenario

**Feature**: skip force-update when `origin/<WtBranch>` tip is not in the local branch

```
# remote-only commit on origin/<WtBranch>
  -> wrk --merge-back -y --push
  -> main pushed; origin/<WtBranch> unchanged; warning; exit 0
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupNotIncludedOriginBranch(t, req)
	req.Args = []string{"--merge-back", "-y", "--push"}
	return nil
}
```
