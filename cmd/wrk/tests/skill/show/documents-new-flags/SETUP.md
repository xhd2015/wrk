# Scenario

**Feature**: embedded SKILL.md documents new exclusive modes including --pr

```
# skill show surfaces agent-facing docs for polish flags
workspace/ -> wrk skill --show
  -> stdout embeds SKILL.md mentioning --propagate-tags, --projects-dep-graph,
     and --pr / --title / --comment
```

## Steps

1. Run `wrk skill --show` from neutral cwd.
2. Assert embedded skill text names the new flags (coverage + P3 surface polish).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show"}
	return nil
}
```
