# Scenario

**Feature**: binary writes follow-up `cd` lines when WRK_FOLLOWUP_FILE is set

```
WRK_FOLLOWUP_FILE=tmp wrk <mode> -> file contains 0..1 lines: cd /abs
```

## Preconditions

- Git required for create/done/set-task success paths.

## Steps

1. Set Mode to binary; prepare empty follow-up file path.
2. Descendants configure CLI args, env, and git fixtures.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.Mode = "binary"
	if req.FollowupFile == "" {
		req.FollowupFile = defaultFollowupPath(req)
	}
	return nil
}
```
