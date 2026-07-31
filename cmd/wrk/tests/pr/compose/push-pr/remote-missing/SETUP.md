# Scenario

**Feature**: `--push --pr` when remote head is missing full-pushes then creates PR

```
# feature only local; origin has no refs/heads/feature-pr
linked wt (feature-pr) + github origin (main only) + fake gh list=[]
  -> wrk --pr --title "Fix login" --comment "please review" --push
  -> full push creates origin/feature-pr
  -> then gh pr create (body = comment)
  -> stdout: pushed … then PR created block
```

## Steps

1. Seed linked feature with commit **not** on origin.
2. Install fake gh (empty list).
3. Run `--pr … --push` (argv order free; execution still push then pr).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeature(t, req)
	installFakeGh(t, req)
	// Prove flag order free: --pr before --push.
	req.Args = prPrThenPushArgs()
	return nil
}
```
