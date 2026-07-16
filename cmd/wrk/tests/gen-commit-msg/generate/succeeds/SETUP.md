# Scenario

**Feature**: wrk --gen-commit-msg with fake-opencode prints mock title and description

```
# stage 1 text file; agent path; no --commit
repo/ (1 staged) -> wrk --gen-commit-msg --agent-runner opencode --agent-runner-binary <fake> --model openai/gpt-5
  -> exit 0
  -> stdout contains "feat: add feature" and "Implement feature X"
  -> HEAD subject unchanged (no commit)
```

## Preconditions

- Isolated git repo with hooks disabled; one staged text file.
- FAKE_OPENCODE_MOCK_CONFIG returns JSON commit message (title + description).
- OPENCODE_CONFIG_DIR points at a WorkRoot temp dir.

## Steps

1. Stage `change.go` via `stageOneTextFile`.
2. Write mock config (`sess_commit` / feat: add feature).
3. Install fake-opencode ExtraEnv.
4. Run wrk with agent-runner flags; no `--commit`.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	writeMockConfig(t, req, mockConfigAddFeature)
	installFakeOpencodeEnv(t, req)
	req.Args = genCommitMsgAgentArgs(req)
	return nil
}
```
