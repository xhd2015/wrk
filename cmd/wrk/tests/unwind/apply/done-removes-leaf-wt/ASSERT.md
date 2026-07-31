## Expected

- Exit code 0 (RED while apply stubbed).
- Leaf external worktree path (`req.DepsLinkedWtDir`) **does not exist**.
- Pin still applied on root main: require `example.com/dot-pkgs` at `v0.0.2`.
- Leaf main advanced + tagged (land before remove).

## Side Effects

- `--done` removes linked leaf WT after successful land; pin uses leaf **main** dir.

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
	if resp.ExitCode != 0 {
		combined := resp.Stdout + "\n" + resp.Stderr
		if strings.Contains(combined, "not implemented") {
			t.Fatalf("apply not implemented yet (expected RED until P4 lands): exit=%d stderr=%q stdout=%q",
				resp.ExitCode, resp.Stderr, resp.Stdout)
		}
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.DepsLinkedWtDir == "" {
		t.Fatal("DepsLinkedWtDir must be set")
	}
	assertFileNotExists(t, req.DepsLinkedWtDir)
	assertConsumerPinned(t, req)
	assertLeafMainAdvancedAndTagged(t, req)
}
```
