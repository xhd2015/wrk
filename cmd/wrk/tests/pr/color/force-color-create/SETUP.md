# Scenario

**Feature**: new-PR success tokens are green under `--color`

```
# remote head exists; no open PR; --color forces ANSI on non-TTY
linked wt + origin/feature-pr + fake gh (list empty)
  -> wrk --pr --title "Fix login" --comment "please review" --color
  -> green: PR created / title set / comment added
  -> plain URL line
```

## Steps

1. Seed linked feature with remote head present (no ensure-push line).
2. Install fake gh (empty list).
3. Run bare `--pr` with `--color`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExistsLocalAhead(t, req)
	installFakeGh(t, req)
	req.Args = append(prDefaultArgs(), "--color")
	return nil
}
```
