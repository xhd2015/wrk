# Scenario

**Feature**: negative flags clear/disable create axes in config

```
wrk --set-config --create --no-open-in-agent | --no-new-window
  -> agent.enabled=false | window cleared/off
```

## Steps

- Leaves seed enabled config then run negative set-config.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
