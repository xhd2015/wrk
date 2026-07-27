# Scenario

**Feature**: Pass 2 distribute — linked worktree ← main (FF-only)

```
# for each linked named-branch wt:
#   if wt dirty / not strictly FF-behind main → skip (+warning when applicable)
#   else git -C wt merge --ff-only <mainBranch>
main ahead of wt -> wrk --sync -> pass2 distribute or skip
```

## Preconditions

- Cwd is the main checkout on branch `main`.
- Pass 2 runs after pass 1; leaves here isolate distribute by having main ahead
  (and wt not ahead) so pass 1 is a silent no-op.

## Steps

- Descendants build the fixture and set `req.Args` / SHA fields.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
