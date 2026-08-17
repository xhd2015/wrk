## Expected Output

```
dep-update example.com/dep -> v0.0.2
go mod tidy ok  module example.com/app
dep-update example.com/dep -> v0.0.2
go mod tidy ok  module example.com/app/pkg
```

## Expected

- Exit 0.
- Per consumer: pin line then tidy line (scan order: `.` then `pkg/`).
- Both go.mods: replace dropped; require `example.com/dep` @ `v0.0.2`.
- Both consumers have go.sum.
- No dir-mode summary.

## Side Effects

- Two existing requirers pinned and tidied; no new requires added elsewhere.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertNotContains(t, resp.Stdout, "would:")
	assertDepUpdateLine(t, resp.Stdout, modDep, req.WantVersion)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumerModule)
	assertTidyOkLine(t, resp.Stdout, req.WantConsumer2Module)
	iApp := strings.Index(resp.Stdout, "go mod tidy ok  module "+req.WantConsumerModule+"\n")
	iPkg := strings.Index(resp.Stdout, "go mod tidy ok  module "+req.WantConsumer2Module)
	if iApp < 0 || iPkg < 0 || iApp > iPkg {
		t.Fatalf("expected tidy app then pkg; got:\n%s", resp.Stdout)
	}
	assertNotContains(t, resp.Stdout, "dep-update: updated")
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertNoReplaceFor(t, req.Consumer2GoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.Consumer2GoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertGoSumExists(t, req.Consumer2ModDir)
}
```
