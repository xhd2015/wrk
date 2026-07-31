# Scenario

**Feature**: when origin lacks the head branch, `--pr` ensures push then creates PR

```
# feature only local; origin has no refs/heads/feature-pr
linked wt (feature-pr) + github origin (main only)
  -> wrk --pr --title "Fix login" --comment "please review"
  -> git push HEAD to origin/feature-pr
  -> stdout starts with: pushed feature-pr → origin/feature-pr
  -> gh pr create + comment
  -> PR created / title set / comment added / URL
```

## Steps

1. Seed main+github origin; linked feature with commit (not pushed).
2. Install fake `gh` (empty list → create path).
3. Run `wrk --pr --title … --comment …` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeature(t, req)
	installFakeGh(t, req)
	req.Args = prDefaultArgs()
	return nil
}
```
