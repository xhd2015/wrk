# Scenario

**Feature**: TTY named bring with multiple linked WTs of source reuses lex-smallest on default skip

```
# two WRK_HOME worktrees of myrepo (suffix 0 and -1)
# TTY + Enter -> skip; stdout = lex-smallest; multi/also-present style warnings ok
myrepo-main-{date} + myrepo-main-{date}-1
  -> wrk myrepo {WorkRoot}/target (\n)
  -> stdout smallest path; no new under target
```

## Steps

1. Pre-create two linked worktrees of `myrepo` via sequential bare `wrk`.
2. Run named bring under fake TTY with stdin `\n`.

```go
func Setup(t *testing.T, req *Request) error {
	paths := namedBringExistingWorktrees(t, req, 2)
	// Lex-smallest of the two abs paths is the unsuffixed one for this naming scheme.
	req.WtDir = paths[0]
	req.ExternalWtDir2 = paths[1] // reuse field for second existing path
	req.StdinInput = "\n"
	return nil
}
```
