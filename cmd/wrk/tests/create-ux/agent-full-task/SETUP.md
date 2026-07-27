# Scenario

**Feature**: agent prompt always receives full original taskDesc even when path/branch slug was budget-trimmed

```
wrk -t <full long task> --open-in-agent  (long basename repo)
  -> create with fitted names ≤255
  -> agent-run last argv = /brainstorm <full taskDesc>  (not slug-only)
```

## Preconditions

- Fake `agent-run` + create-ux mocks installed (inherited from create-ux root helpers).
- Name budget fit shortens slug for path/branch; agent must not receive fitted slug as task text.

## Steps

- Leaves set long basename repo, long TaskDesc, `--open-in-agent`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
