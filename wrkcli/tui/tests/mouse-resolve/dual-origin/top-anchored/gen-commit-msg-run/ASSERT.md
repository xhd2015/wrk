## Expected

- `err` is nil.
- Resolve hits (`resp.OK`).
- `resp.RunStage == "gen-commit-msg"`.
- `resp.RunStage` is **not** `"tag-next"` (bug regression).
- Hitmap contains both gen-commit-msg and tag-next Run regions at different Y
  (sanity: wrong mapping would still have a tag-next target to confuse with).

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	var genY, tagY int
	var haveGen, haveTag bool
	for _, h := range resp.Hits {
		switch h.RunStage {
		case "gen-commit-msg":
			haveGen = true
			genY = h.Y0
		case "tag-next":
			haveTag = true
			tagY = h.Y0
		}
	}
	if !haveGen || !haveTag {
		t.Fatalf("hitmap must include gen-commit-msg and tag-next Run hits; haveGen=%v haveTag=%v hits=%d",
			haveGen, haveTag, len(resp.Hits))
	}
	if genY == tagY {
		t.Fatalf("gen-commit-msg and tag-next Run must differ in local Y; both y0=%d", genY)
	}
	if !resp.OK {
		t.Fatalf("expected hit on gen-commit-msg Run; miss (absX=%d absY=%d height blank=%d viewLines=%d aimedLocalY=%d)",
			resp.AimedAbsX, resp.AimedAbsY, req.ExtraBlank, resp.ViewLines, resp.AimedLocalY)
	}
	if resp.RunStage == "tag-next" {
		t.Fatalf("BUG: top-anchored gen-commit-msg Run resolved to tag-next (localY=%d originKind=%q absY=%d viewLines=%d)",
			resp.LocalY, resp.OriginKind, resp.AimedAbsY, resp.ViewLines)
	}
	if resp.RunStage != "gen-commit-msg" {
		t.Fatalf("runStage: want gen-commit-msg, got %q (localY=%d originKind=%q focus=%d)",
			resp.RunStage, resp.LocalY, resp.OriginKind, resp.Focus)
	}
}
```
