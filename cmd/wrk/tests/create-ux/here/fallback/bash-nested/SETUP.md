# Scenario

**Feature**: `--here` fallback nests bash at worktree with agent startup on rcfile

```
SHELL=bash (fake); no WRK_FOLLOWUP_FILE
wrk -t 'here fallback' --open-in-agent --here
  -> stdout worktree
  -> stderr install hint
  -> fake bash cwd = worktree; args include --rcfile
  -> outer agent-run not invoked (fake bash does not source rc)
```

## Steps

1. Install fake bash; run `--here --open-in-agent`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	installFakeBashUX(t, req, 0)
	req.TaskDesc = "here fallback"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--here"}
	return nil
}
```
