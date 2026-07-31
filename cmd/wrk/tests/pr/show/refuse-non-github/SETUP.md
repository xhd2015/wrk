# Scenario

**Feature**: bare `wrk --pr` show mode refuses when origin is not github.com

```
# origin is local bare path (not github.com)
linked wt + bare origin
  -> wrk --pr
  -> non-zero
  -> stderr explains github / origin requirement
  -> no gh create/comment
```

## Steps

1. Seed linked feature with non-github bare origin.
2. Install fake gh.
3. Run bare `--pr` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrNonGithubOrigin(t, req)
	installFakeGh(t, req)
	req.Args = prShowArgs()
	return nil
}
```
