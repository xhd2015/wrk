# Scenario

**Feature**: embedded SKILL.md documents --propagate-tags and --projects-dep-graph

```
# skill show surfaces agent-facing docs for new exclusive modes
workspace/ -> wrk skill --show
  -> stdout embeds SKILL.md mentioning --propagate-tags and --projects-dep-graph
```

## Steps

1. Run `wrk skill --show` from neutral cwd.
2. Assert embedded skill text names the new flags (P8 polish).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show"}
	return nil
}
```
