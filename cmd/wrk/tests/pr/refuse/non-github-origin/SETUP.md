# Scenario

**Feature**: `--pr` refuses when origin is not github.com

```
# origin is local bare path (not github.com)
linked wt + bare origin
  -> wrk --pr --title T --comment C
  -> non-zero
  -> stderr explains github / origin requirement
```

## Steps

1. Seed linked feature with non-github bare origin.
2. Install fake gh.
3. Run default `--pr` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrNonGithubOrigin(t, req)
	installFakeGh(t, req)
	req.Args = prDefaultArgs()
	return nil
}
```
