# Scenario

**Feature**: WhereMain on a main checkout returns the cleaned self path

```
# main only
Caller -> WhereMain(mainRepo) -> mainAbs == cleaned mainRepo
```

## Steps

1. Seed a main repository (no linked worktree).
2. Set Checkout to main.
3. Run WhereMain.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMainRepo(t, req, "myrepo")
	req.Checkout = req.MainRepo
	return nil
}
```
