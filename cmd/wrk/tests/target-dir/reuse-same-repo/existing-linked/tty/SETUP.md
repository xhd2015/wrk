# Scenario

**Feature**: named bring Policy B on a fake TTY (`script`) with stdin answers

```
# UseScriptTTY + StdinInput drives Y/n prompt on stderr
existing linked WT of myrepo + TTY
  -> prompt: <basename> already exists in <path>, skip? [Y/n]
  -> answer Y/empty -> skip; answer n -> create under target
```

## Steps

- Descendants set `req.UseScriptTTY = true` and `req.StdinInput`.

```go
func Setup(t *testing.T, req *Request) error {
	// Fake TTY via root Run's execScriptTTYWrk path.
	req.UseScriptTTY = true
	if !req.UseScriptTTY {
		t.Fatal("UseScriptTTY must be true for Policy B TTY leaves")
	}
	return nil
}
```
