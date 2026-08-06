# Scenario

**Feature**: non-TTY named create with reusable same-parent sibling creates (does not refuse)

```
# no fake TTY; piped/non-interactive stdin
# new policy: create as today — NOT hard refuse
reusable sibling under same parent + non-TTY
  -> wrk myrepo {WorkRoot}/target
  -> exit 0; new path; no refusing non-interactive
```

## Steps

- Descendants pre-create reusable siblings; leave `UseScriptTTY` false (default pipe).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseScriptTTY = false
	if req.UseScriptTTY {
		t.Fatal("UseScriptTTY must be false for non-TTY create leaves")
	}
	return nil
}
```
