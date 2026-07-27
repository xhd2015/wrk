## Expected

- Stderr does NOT contain `failed to find merge base` (the cascade must not
  crash comparing the dep worktree branch against consumer main).
- External dependency worktree under `external/` no longer exists (the cascade
  removes it).

## Exit Code

- Not asserted by this leaf. The merge-base crash is the sole focus; the
  consumer worktree is left dirty by `wrk --dep` + the `dropreplace`, so its own
  merge-back fate is out of scope here (covered by sibling leaves). The two
  assertions above are sufficient to prove the cascade bug: today both fail
  (stderr contains the merge-base error and the ext wt is left in place); once
  the cascade removes the dep worktree without crashing, both pass.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)

	assertNotContains(t, resp.Stderr, "failed to find merge base")
	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.MainRepo, req.ExternalWtDir)
}
```
