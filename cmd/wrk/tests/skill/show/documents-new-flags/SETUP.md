# Scenario

**Feature**: embedded SKILL.md documents polish flags and multi-mode `--pr`

```
# skill show surfaces agent-facing docs for polish flags + multi-mode PR
workspace/ -> wrk skill --show
  -> stdout embeds SKILL.md mentioning --propagate-tags, --projects-dep-graph,
     and multi-mode --pr (show / status / comment / create; --title/--comment
     for create-attach — not always-required companions)
```

## Steps

1. Run `wrk skill --show` from neutral cwd.
2. Assert embedded skill text names the polish flags and multi-mode PR contract
   (coverage + P5 surface polish).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show"}
	return nil
}
```
