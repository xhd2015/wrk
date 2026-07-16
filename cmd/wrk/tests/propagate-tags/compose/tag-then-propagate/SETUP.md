# Scenario

**Feature**: compose apply creates next source tag then bumps and commits consumer

```
# lib tagged v1.0.0 + post-tag owned change; app requires example.com/lib@v1.0.0
cwd=lib -> wrk --tag-next --propagate-tags
  -> (1) tag-next: create v1.0.1 at HEAD; "1 tag created"
  -> blank line
  -> (2) propagate: source @ v1.0.1; update app v1.0.0 -> v1.0.1
       go build ./... ok; committed chore(deps): …
  -> app go.mod require becomes v1.0.1; app HEAD advances
  -> source HEAD unchanged (tag only); source go.mod unchanged
```

## Steps

1. `setupComposeRootBump` (no origin).
2. Args: `--tag-next --propagate-tags` (flag order free; this leaf uses tag-next first).

```go
func Setup(t *testing.T, req *Request) error {
	setupComposeRootBump(t, req, false)
	req.Args = []string{"--tag-next", "--propagate-tags"}
	return nil
}
```
