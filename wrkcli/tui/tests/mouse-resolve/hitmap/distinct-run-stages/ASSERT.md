## Expected

- Hitmap includes Run hits for both `gen-commit-msg` and `tag-next`.
- Their local `Y0` values differ.
- `viewLines > 0`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("hitmap: %v", err)
	}
	if resp.ViewLines <= 0 {
		t.Fatalf("viewLines must be > 0, got %d", resp.ViewLines)
	}
	var genY, tagY int
	var genX0, genX1, tagX0, tagX1 int
	var haveGen, haveTag bool
	for _, h := range resp.Hits {
		switch h.RunStage {
		case "gen-commit-msg":
			haveGen = true
			genY, genX0, genX1 = h.Y0, h.X0, h.X1
		case "tag-next":
			haveTag = true
			tagY, tagX0, tagX1 = h.Y0, h.X0, h.X1
		}
	}
	if !haveGen {
		t.Fatal("missing runStage hit for gen-commit-msg")
	}
	if !haveTag {
		t.Fatal("missing runStage hit for tag-next")
	}
	if genY == tagY {
		t.Fatalf("gen-commit-msg and tag-next Run must have different local Y; both y0=%d", genY)
	}
	// Run chips should be on the right (non-empty x range)
	if genX1 <= genX0 || tagX1 <= tagX0 {
		t.Fatalf("Run hits need x1>x0; gen=[%d,%d) tag=[%d,%d)", genX0, genX1, tagX0, tagX1)
	}
}
```
