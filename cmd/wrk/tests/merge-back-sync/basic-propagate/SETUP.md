# Scenario

**Feature**: after `--merge-back --sync`, remaining behind worktree gets main tip; wtA kept

```
# wtA ahead; wtB behind → merge-back keeps wtA; sync pass2 to feature-stays
myrepo + wtA (ahead) + wtB (feature-stays)
  -> wrk --merge-back -y --sync
  -> merged branch <wtA> into main
  -> <blank>
  -> feature-stays ← main  (+1 commit)
  -> synced: 0 into main, 1 into worktrees, 0 skipped
  -> wtA still present; no "worktree removed:"
  -> wtB HEAD == main HEAD
```

## Steps

1. Create wrk-managed wtA + manual linked `feature-stays` wtB; commit ahead on wtA.
2. Run `wrk --merge-back -y --sync` from wtA.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCompositionTwoWTs(t, req)
	req.Args = []string{"--merge-back", "-y", "--sync"}
	return nil
}
```
