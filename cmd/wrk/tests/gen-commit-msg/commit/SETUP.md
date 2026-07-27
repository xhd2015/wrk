# Scenario

**Feature**: wrk --gen-commit-msg --commit via fake-opencode

```
# agent generates message then git commit (optional --no-verify)
staged + fake-opencode
  -> wrk --gen-commit-msg ... --commit [--no-verify]
  -> HEAD subject = mock title
```

## Preconditions

- Root harness has WorkRoot / WrkHome.
- Session fake-opencode is buildable from external agent-pro.
- Leaves stage a git repo (hooks disabled unless no-verify leaf).

## Steps

1. Inherit root Setup.
2. Leaf stages files, writes mock config, installs fake-opencode env, sets Args with `--commit`.

## Context

- `commit/succeeds` uses hooks-disabled repo and mock title `feat: add feature`.
- `commit/no-verify` uses failing pre-commit + mock title `feat: skip hooks`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: commit leaves share fake-opencode helpers.
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
