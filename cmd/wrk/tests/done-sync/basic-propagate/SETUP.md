# Scenario

**Feature**: after `--done --sync`, remaining behind worktree receives main tip (pass2)

```
# wtA ahead; wtB at old base → done removes wtA; sync pass2: feature-stays ← main (+1)
myrepo + wtA (ahead) + wtB (feature-stays)
  -> wrk --done -y --sync
  -> merged branch <wtA> into main
  -> <blank>
  -> feature-stays ← main  (+1 commit)
  -> synced: 0 into main, 1 into worktrees, 0 skipped
  -> wtA gone; wtB HEAD == main HEAD
```

## Steps

1. Create wrk-managed wtA + manual linked `feature-stays` wtB at shared tip.
2. Commit ahead on wtA (`feature-work`).
3. Run `wrk --done -y --sync` from wtA.

```go
func Setup(t *testing.T, req *Request) error {
	setupCompositionTwoWTs(t, req)
	req.Args = []string{"--done", "-y", "--sync"}
	return nil
}
```
