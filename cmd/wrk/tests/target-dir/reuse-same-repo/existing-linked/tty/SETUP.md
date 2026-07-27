# Scenario

**Feature**: named bring Policy B on a fake TTY (`script`) with stdin answers

```
# UseScriptTTY + StdinInput drives Y/n prompt on stderr
existing linked WT of myrepo + TTY
  -> prompt: wrk: warning: <basename> already has a linked worktree at <path>, skip creating another? [Y/n]
  -> answer Y/empty -> skip; answer n -> create under target
```

## Steps

- Descendants set `req.UseScriptTTY = true` and `req.StdinInput`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Fake TTY via root Run's execScriptTTYWrk path.
	req.UseScriptTTY = true
	if !req.UseScriptTTY {
		t.Fatal("UseScriptTTY must be true for Policy B TTY leaves")
	}
	return nil
}
```
