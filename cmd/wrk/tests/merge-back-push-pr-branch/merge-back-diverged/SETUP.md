# Scenario

**Feature**: `--merge-back --push` lease-updates `origin/<WtBranch>` to the rebased tip

```
# main and feature diverged; origin/<WtBranch> still has pre-rebase tip
  -> wrk --merge-back -y --push
  -> rebased and merged; origin/<WtBranch> == new main HEAD
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDivergedWithOriginBranch(t, req)
	req.Args = []string{"--merge-back", "-y", "--push"}
	return nil
}
```
