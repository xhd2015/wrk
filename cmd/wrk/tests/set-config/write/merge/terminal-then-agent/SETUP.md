# Scenario

**Feature**: terminal-only write then agent-only write keeps both

```
1) --set-config --create --new-terminal
2) --set-config --create --open-in-agent
  -> terminal.mode=new still present; agent.enabled=true
```

## Steps

1. Run terminal-only set-config (via helper; leaf Run is second write).
2. Leaf Run: agent-only set-config.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	first := runWrkSetConfig(t, req, setConfigArgs("--create", "--new-terminal")...)
	if first.ExitCode != 0 {
		t.Fatalf("first set-config exit %d stderr=%q", first.ExitCode, first.Stderr)
	}
	req.Args = setConfigArgs("--create", "--open-in-agent")
	return nil
}
```
