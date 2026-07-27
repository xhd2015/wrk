# Scenario

**Feature**: named bring with existing linked WT and non-TTY stdin hard-refuses create

```
# no fake TTY; piped/non-interactive stdin
existing linked WT of myrepo
  -> wrk myrepo <target-dir>
  -> non-zero; stderr refuse message; empty stdout; no new WT
```

## Steps

- Descendants pre-create linked WTs; leave `UseScriptTTY` false (default pipe).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Explicit non-TTY: no UseScriptTTY; no confirm escape hatches.
	req.UseScriptTTY = false
	if req.UseScriptTTY {
		t.Fatal("UseScriptTTY must be false for non-TTY refuse leaves")
	}
	return nil
}
```
