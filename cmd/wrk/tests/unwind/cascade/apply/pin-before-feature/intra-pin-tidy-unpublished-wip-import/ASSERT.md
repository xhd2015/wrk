## Expected

- Exit code 0.
- Combined stdout/stderr must **not** contain `does not contain package` or
  `unknown revision`.
- Free monorepo after apply (`req.SecondRepo`):
  - `shell/openterm2/openterm2.go` present on free main (unpublished WIP landed)
  - local root tag `v0.0.2` exists
- Consumer after apply:
  - `require example.com/dot-pkgs v0.0.2`
  - **no** droppable external replace for free root
  - **must not** require `example.com/dot-pkgs/cmd`
  - intra `example.com/aaa-spl/tools/trace/trace_tool` require is `v0.0.16`
- Cascade pin of free root @ next present when require bumped.

## Side Effects

- Today RED: cascade intra pin tidies Base go.mod (WIP replace stripped) while
  WT source imports unpublished `shell/openterm2` → tidy exit 1 /
  `@latest … does not contain package`.
- Desired: free feature lands and is tagged next before any consumer tidy that
  resolves free via the network; intra pin must not drop the replace (or must
  wait) until that version exists.

## Errors

- None on success path.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	out := resp.Stdout + "\n" + resp.Stderr
	if resp.ExitCode != 0 {
		t.Fatalf("CS-openterm2: want exit 0 (intra pin tidy must not die on unpublished import); exit=%d\nstderr=%q\nstdout=%q\ncombined:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout, out)
	}
	if req.SecondRepo == "" || req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("SecondRepo, MainRepo, and LeafModulePath required")
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "does not contain package") {
		t.Fatalf("CS-openterm2: tidy must not report missing unpublished package\ncombined:\n%s", out)
	}
	if strings.Contains(lower, "unknown revision") {
		t.Fatalf("CS-openterm2: unknown revision after apply\ncombined:\n%s", out)
	}

	openPath := filepath.Join(req.SecondRepo, "shell", "openterm2", "openterm2.go")
	if _, err := os.Stat(openPath); err != nil {
		t.Fatalf("CS-openterm2: free main missing landed shell/openterm2: %v", err)
	}
	if !tagRefExists(t, req.SecondRepo, req.ExpectedPinVersion) {
		t.Fatalf("CS-openterm2: free main missing root tag %s\nlocal tags:\n%s",
			req.ExpectedPinVersion, gitOutputIsolated(t, req.SecondRepo, "tag", "-l"))
	}

	assertConsumerRequireAndNoExternalReplace(t, req)
	assertConsumerPinnedRootOnly(t, req)

	checkout := req.MainRepo
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			checkout = req.WtDir
		}
	}
	gotTool := requireVersionInGoMod(t, filepath.Join(checkout, "go.mod"), csOpenterm2ToolModule)
	if gotTool != csOpenterm2ToolNextVer {
		t.Fatalf("CS-openterm2: consumer require %s = %q, want %s\ngo.mod:\n%s",
			csOpenterm2ToolModule, gotTool, csOpenterm2ToolNextVer,
			readFile(t, filepath.Join(checkout, "go.mod")))
	}

	hist := historyRepoForConsumer(t, req)
	assertCascadePinCommitPresent(t, hist, req.LeafModulePath, req.ExpectedPinVersion)
	assertGoModCommittedClean(t, hist)
}
```
