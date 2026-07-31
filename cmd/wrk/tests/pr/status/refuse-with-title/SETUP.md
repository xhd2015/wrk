# Scenario

**Feature**: `--pr --status` cannot combine with `--title` / `--comment`

```
# linked wt; invalid combination
linked wt + github origin + gh
  -> wrk --pr --status --title "Fix login" --comment "please review"
  -> non-zero
  -> stderr indicates invalid combination / mutual exclusion / not valid
  -> no gh pr create / status view success path
```

## Steps

1. Seed linked feature with remote head (gates would pass if status alone were allowed).
2. Install fake gh (unused on early refuse).
3. Run `--pr --status --title T --comment C`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	installFakeGh(t, req)
	req.Args = []string{"--pr", "--status", "--title", prDefaultTitle, "--comment", prDefaultComment}
	return nil
}
```
