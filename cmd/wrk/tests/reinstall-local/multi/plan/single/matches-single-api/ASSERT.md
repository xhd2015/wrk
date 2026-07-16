## Expected

- `err` is nil.
- Multi plan has one module with ModuleName `multi-single` and one install item
  for bin `present` (`go-install`, `./cmd/present`).
- That module's Items equal `PlanLocalReinstalls(moduleRoot, binDir).Items`
  (BinName, Method, RelPath, Action) — single-module equivalence.

## Side Effects

- None beyond pure reads of fixtures already written by Setup.

## Exit Code

- N/A (no process).

```go
import (
	"reflect"
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertMultiPlanOK(t, req, resp, err)
	if len(req.ModuleRoots) != 1 {
		t.Fatalf("test fixture: want 1 ModuleRoot, got %d", len(req.ModuleRoots))
	}
	single, serr := wrkcli.PlanLocalReinstalls(req.ModuleRoots[0], req.BinDir)
	if serr != nil {
		t.Fatalf("PlanLocalReinstalls (equivalence baseline): %v", serr)
	}
	if single == nil {
		t.Fatal("PlanLocalReinstalls returned nil plan")
	}
	if len(resp.Modules) != 1 {
		t.Fatalf("multi Modules len: got %d want 1", len(resp.Modules))
	}
	got := resp.Modules[0].Items
	want := make([]WantPlanItem, len(single.Items))
	for i, it := range single.Items {
		want[i] = WantPlanItem{
			BinName: it.BinName,
			Method:  string(it.Method),
			RelPath: it.RelPath,
			Action:  string(it.Action),
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multi items ≠ single PlanLocalReinstalls items\n multi: %#v\nsingle: %#v", got, want)
	}
}
```
