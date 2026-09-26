# Scenario

**Feature**: `wrk --add-all --gen-commit-msg --commit` creates the first commit on unborn HEAD

```
# git init only; untracked file; add-all + generate + commit
unborn repo (untracked) -> wrk --add-all --gen-commit-msg ... --commit
  -> exit 0
  -> HEAD subject = "feat: add feature" (root commit)
```

## Preconditions

- Isolated git repo after `git init` (no commits); untracked `change.go`.
- FAKE_OPENCODE_MOCK_CONFIG returns JSON commit message (title + description).

## Steps

1. Init unborn repo with untracked `change.go` (not staged).
2. Write mock config (`sess_commit` / feat: add feature).
3. Install fake-opencode ExtraEnv.
4. Run wrk with `--add-all` and `--commit`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeUntrackedInUnbornRepo(t, req)
	writeMockConfig(t, req, mockConfigAddFeature)
	installFakeOpencodeEnv(t, req)
	req.Args = genCommitMsgAgentArgs(req, "--add-all", "--commit")
	return nil
}
```
