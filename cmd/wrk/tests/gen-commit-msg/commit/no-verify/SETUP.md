# Scenario

**Feature**: wrk --gen-commit-msg --commit --no-verify skips failing pre-commit hook

```
# failing pre-commit + --commit --no-verify through wrk binary
repo/ (hook exits 1) + staged
  -> wrk --gen-commit-msg ... --commit --no-verify
  -> exit 0
  -> HEAD subject = "feat: skip hooks"
```

## Preconditions

- Git repo with a pre-commit hook that always exits 1 (hooks not disabled).
- One staged text file.
- FAKE_OPENCODE_MOCK_CONFIG returns title `feat: skip hooks` (agent-pro no-verify leaf shape).

## Steps

1. Stage change in repo with failing pre-commit via `stageOneTextFileWithFailingPreCommit`.
2. Write mock config (`sess_no_verify` / feat: skip hooks).
3. Install fake-opencode ExtraEnv.
4. Run wrk with `--commit --no-verify`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFileWithFailingPreCommit(t, req)
	writeMockConfig(t, req, mockConfigSkipHooks)
	installFakeOpencodeEnv(t, req)
	req.Args = genCommitMsgAgentArgs(req, "--commit", "--no-verify")
	return nil
}
```
