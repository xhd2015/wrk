# Scenario

**Feature**: TTY Policy B skip prompt for reusable same-parent sibling(s)

```
# UseScriptTTY + StdinInput drives Y/n on stderr
# wrk: warning: … would reuse <path> …
# wrk: warning: … skip creating …? [Y/n]
reusable sibling + TTY
  -> answer Y/empty -> skip; answer n -> create
```

## Steps

- Descendants set `req.UseScriptTTY = true` and `req.StdinInput`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseScriptTTY = true
	if !req.UseScriptTTY {
		t.Fatal("UseScriptTTY must be true for Policy B TTY leaves")
	}
	// TTY path is binary e2e via script(1); not InProcess.
	req.InProcess = false
	return nil
}
```
