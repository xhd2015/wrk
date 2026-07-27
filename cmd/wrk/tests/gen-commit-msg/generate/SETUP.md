# Scenario

**Feature**: wrk --gen-commit-msg generate path via fake-opencode (no --commit)

```
# agent path through wrk binary; mock LLM events; print message only
staged + fake-opencode
  -> wrk --gen-commit-msg --agent-runner opencode --agent-runner-binary <fake> --model openai/gpt-5
  -> stdout: title + description from mock; exit 0; no commit
```

## Preconditions

- Root harness has WorkRoot / WrkHome.
- Session fake-opencode is buildable from external agent-pro.
- Leaves stage a git repo and install mock env.

## Steps

1. Inherit root Setup.
2. Leaf stages files, writes mock config, installs fake-opencode env, sets Args.

## Context

- No `--commit`: HEAD must not move.
- Mock title/description match agent-pro `commit-with-fake-opencode/succeeds`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: generate leaves share fake-opencode helpers.
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
