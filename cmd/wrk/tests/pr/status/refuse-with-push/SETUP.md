# Scenario

**Feature**: `--pr --status` cannot combine with `--push`

```
# linked wt; invalid combination
linked wt + github origin + gh
  -> wrk --pr --status --push
  -> non-zero
  -> stderr indicates invalid combination / mutual exclusion / not valid
  -> no full tip push; origin tip unchanged
```

## Steps

1. Seed linked feature with remote head + local ahead (origin snapshot for push side-effect check).
2. Install fake gh; open PR optional (should refuse before push).
3. Run `--pr --status --push`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// Local ahead + origin snapshot: if product wrongly pushed, origin would advance.
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	setFakeGhExistingPR(t, req, prExistingTitle, prDefaultURL, prExistingNumber)
	req.Args = []string{"--pr", "--status", "--push"}
	return nil
}
```
