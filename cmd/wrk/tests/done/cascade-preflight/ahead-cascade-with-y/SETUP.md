# Scenario

**Feature**: cascade target ahead of its main + `-y` → auto-yes cascade merge + own complete (D3)

```
# external ahead; clean own (replace dropped)
  -> wrk --done -y
  -> cascade merges dep + removes external
  -> own merge-back completes; consumer removed
  -> exit 0
```

## Steps

1. Build ahead external + drop replace via `setupCascadePreflightAheadExternal`.
2. Run `wrk --done -y` (non-TTY pipe; `-y` still auto-yes for cascade).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadePreflightAheadExternal(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
