# Scenario

**Feature**: `--merge-back -y --sync --tag-next` runs sync then local tag-next; source wt kept

```
# wtA ahead; wtB behind; after merge-back: sync pass2 then local v0.0.2
myrepo (v0.0.1) + wtA + wtB (feature-stays)
  -> wrk --merge-back -y --sync --tag-next
  -> merged branch <wtA> into main
  -> <blank>
  -> feature-stays ← main (+1); synced summary
  -> <blank>
  -> tag-next apply v0.0.2 local
  -> no push line; wtA stays; wtB HEAD == main
```

## Steps

1. Root-bump seed; wrk wtA + manual wtB `feature-stays`; commit ahead on wtA.
2. Run `wrk --merge-back -y --sync --tag-next` from wtA.

```go
func Setup(t *testing.T, req *Request) error {
	setupMergeBackPipelineSync(t, req)
	req.Args = []string{"--merge-back", "-y", "--sync", "--tag-next"}
	return nil
}
```
