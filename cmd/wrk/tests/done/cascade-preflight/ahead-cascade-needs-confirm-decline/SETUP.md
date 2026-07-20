# Scenario

**Feature**: cascade target HEAD not contained in its main → must confirm; decline / non-TTY path aborts without remove (D3/D4)

```
# external ahead of dep main; clean own (replace dropped)
  -> wrk --done --confirm-from-stdin  (stdin n)
  -> non-zero or clean abort
  -> external NOT removed
  -> default auto-yes must NOT silently merge cascade (D3)
```

## Steps

1. Build ahead external + drop replace via `setupCascadePreflightAheadExternal`.
2. Run `wrk --done --confirm-from-stdin` with `n\n` (decline cascade not-included confirm).
   Product: cascade not-included requires confirm even under default auto-yes; `-y` is the skip.

```go
func Setup(t *testing.T, req *Request) error {
	setupCascadePreflightAheadExternal(t, req)
	req.RepoDir = req.WtDir
	// D3/D4: not-included cascade target uses PromptConfirmPlan / confirm-from-stdin on pipes.
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "n\n"
	return nil
}
```
