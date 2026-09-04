```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertDryRunBanner(t, resp.Stdout)
	assertWouldTidyLine(t, resp.Stdout, modConsumer)
	goRoot := filepath.Join(req.InstallDir, req.WantGoPin)
	assertContains(t, resp.Stdout, "would: go mod tidy  (local git; go="+req.WantGoPin+"; GOROOT="+goRoot+")")
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
