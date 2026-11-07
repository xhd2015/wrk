# Scenario

**Feature**: embedded SKILL.md documents the ship flags agents need (and omits retired ones)

```
# skill show surfaces agent-facing docs for ship stages + multi-mode --pr
workspace/ -> wrk skill --show
  -> stdout embeds SKILL.md mentioning --bring, --merge-back, --projects-dep-graph,
     --dep-replace, and multi-mode --pr (show / status / comment / create;
     --title/--comment for create-attach — not always-required companions)
  -> stdout must NOT mention --unwind or --dep-update (retired from agent docs)
```

## Steps

1. Run `wrk skill --show` from neutral cwd.
2. Assert embedded skill text names the ship flags and no longer advertises
   `--unwind` / `--dep-update`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show"}
	return nil
}
```
