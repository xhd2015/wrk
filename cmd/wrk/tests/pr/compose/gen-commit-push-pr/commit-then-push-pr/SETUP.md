# Scenario

**Feature**: `--gen-commit-msg --commit --push --pr` commits staged work then pushes then opens PR

```
# remote feature exists (so bare --pr would not push); staged uncommitted file
linked wt + origin/feature-pr + staged compose-stage.go + commandcode mock + fake gh
  -> wrk --gen-commit-msg --commit --agent-runner commandcode --agent-runner-binary <mock>
         --push --pr --title "Fix login" --comment "please review"
  -> HEAD subject = "feat: compose pr"
  -> origin tip == new HEAD (full push after commit)
  -> stdout includes commit title, push confirm, PR created block (in that order)
  -> gh pr create (body = comment)
```

## Steps

1. Seed linked feature with remote head present.
2. Disable host hooks (`core.hooksPath=/dev/null`) so `--commit` is hermetic.
3. Stage `compose-stage.go` on the worktree (uncommitted).
4. Install commandcode shell mock + fake gh (empty list).
5. Run gen-commit + push + pr argv from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrLinkedFeatureRemoteExists(t, req)
	// Match gen-commit-msg hermeticity: real commit must not hit host hooksPath.
	disablePrRepoHooks(t, req)
	stagePrComposeFile(t, req)
	agentBin := installCommandCodeCommitMock(t, req)
	installFakeGh(t, req)
	req.Args = prGenCommitPushPrArgs(agentBin)
	return nil
}
```
