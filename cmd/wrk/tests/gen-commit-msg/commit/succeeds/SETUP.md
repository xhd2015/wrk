# Scenario

**Feature**: wrk --gen-commit-msg --commit creates commit with mock title

```
# stage 1 text file; agent path; --commit
repo/ (1 staged) -> wrk --gen-commit-msg --agent-runner opencode --agent-runner-binary <fake> --model openai/gpt-5 --commit
  -> exit 0
  -> HEAD subject = "feat: add feature"
```

## Preconditions

- Isolated git repo with hooks disabled; one staged text file.
- FAKE_OPENCODE_MOCK_CONFIG returns JSON commit message (title + description).

## Steps

1. Stage `change.go` via `stageOneTextFile`.
2. Write mock config (`sess_commit` / feat: add feature).
3. Install fake-opencode ExtraEnv.
4. Run wrk with agent-runner flags and `--commit`.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	writeMockConfig(t, req, mockConfigAddFeature)
	installFakeOpencodeEnv(t, req)
	req.Args = genCommitMsgAgentArgs(req, "--commit")
	return nil
}
```
