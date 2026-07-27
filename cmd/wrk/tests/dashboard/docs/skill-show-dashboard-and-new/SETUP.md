# Scenario

**Feature**: `wrk skill --show` documents dashboard and create via `--new`

```
workspace/ -> wrk skill --show
  -> stdout embeds SKILL.md
  -> mentions dashboard and/or bare no-args not creating
  -> documents wrk --new (or --new create guidance)
```

## Steps

1. Run `wrk skill --show` from neutral cwd.
2. Assert agent-facing skill text reflects create entry change (P4).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"skill", "--show"}
	return nil
}
```
