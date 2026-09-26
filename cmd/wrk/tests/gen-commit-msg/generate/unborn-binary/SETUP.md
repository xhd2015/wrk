# Scenario

**Feature**: `wrk --gen-commit-msg` auto-unstages a binary on unborn HEAD

```
# git init only; staged text + ELF binary; Generate must not fatal restore --staged
unborn repo (app.go + blob.bin staged) -> wrk --gen-commit-msg --agent-runner opencode ...
  -> exit 0
  -> blob.bin unstaged (still on disk); app.go still staged
  -> no "could not resolve 'HEAD'"
```

## Preconditions

- Isolated git repo after `git init` (no commits); staged `app.go` + `blob.bin`.
- FAKE_OPENCODE_MOCK_CONFIG returns JSON commit message.

## Steps

1. Init unborn repo and stage text + binary.
2. Write mock config; install fake-opencode ExtraEnv.
3. Run wrk generate (no `--commit`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageBinaryAndTextFileUnborn(t, req)
	writeMockConfig(t, req, mockConfigAddFeature)
	installFakeOpencodeEnv(t, req)
	req.Args = genCommitMsgAgentArgs(req)
	return nil
}
```
