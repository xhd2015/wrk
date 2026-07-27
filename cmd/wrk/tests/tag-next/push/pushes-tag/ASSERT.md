
## Expected Output

Human tag-next apply block, blank line, then branch push confirmation:

```
v0.0.1        owned changed                  ->  v0.0.2
tagged v0.0.2 @ <short>
1 tag created

pushed main → origin/main
```

## Expected

- Exit code 0.
- Local tag `v0.0.2` exists at HEAD.
- Bare `origin` has `refs/tags/v0.0.2`.
- Bare `origin` `refs/heads/main` equals local main HEAD (branch tip pushed, not tags-only).
- Stdout: tag-next apply human output, blank line, `pushed main → origin/main`.

## Side Effects

- Branch + newly created tags published via `runPushMain(main, tags)` semantics
  (same family as done-pipeline tag-next-push), not tagscope tags-only push.

## Exit Code

- 0

```go
import (
	"fmt"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	if !tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 should exist locally after --push apply")
	}
	if !remoteTagExists(t, req.OriginBare, "v0.0.2") {
		t.Fatal("v0.0.2 should exist on bare origin after --push")
	}

	// Branch tip must be on origin (not tags-only push).
	localMain := strings.TrimSpace(gitOutputIsolated(t, req.MainRepo, "rev-parse", "HEAD"))
	originMain := strings.TrimSpace(gitOutputIsolated(t, req.OriginBare, "rev-parse", "refs/heads/main"))
	if originMain != localMain {
		t.Fatalf("origin/main %s != local HEAD %s (branch must be pushed with tags)", originMain, localMain)
	}

	short := shortHEAD(t, req.MainRepo)
	tagBlock := fmt.Sprintf(
		"v0.0.1        owned changed                  ->  v0.0.2\ntagged v0.0.2 @ %s\n1 tag created\n",
		short,
	)
	want := strings.TrimSuffix(tagBlock, "\n") + "\n\n" + "pushed main → origin/main\n"
	assert.Output(t, resp.Stdout, tagNextStdoutV2(want))
}
```
