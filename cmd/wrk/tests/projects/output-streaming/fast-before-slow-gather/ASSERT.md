---
label: slow
explanation: two tracked repos; zzz has 12 linked worktrees + bare remote; cold run ~35s
---

## Expected

- Exit code 0; stderr empty.
- Full stdout: two lex-ordered detailed blocks (`aaa` then `zzz`), blank line separator.
- **Streaming UX**: first stdout bytes arrive at least 40ms before run completes, and the first chunk is the `aaa` block (fast project printed while `zzz` is still gathering).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"sort"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	streamingProjectsBlocksSeparated(t, resp.Stdout, 2)

	paths := []string{resolvePath(t, req.MainRepo), resolvePath(t, req.SecondRepo)}
	sort.Strings(paths)
	repoByPath := map[string]string{
		resolvePath(t, req.MainRepo):   req.MainRepo,
		resolvePath(t, req.SecondRepo): req.SecondRepo,
	}

	var blocks []string
	for _, abs := range paths {
		repo := repoByPath[abs]
		var remote string
		if filepath.Base(repo) == "aaa" {
			remote = "Remote:       (no upstream)"
		} else {
			remote = streamingCompareWithRemoteField(t, repo, "origin/main", "main")
		}
		var summary string
		if filepath.Base(repo) == "zzz" {
			summary = "12 total, 0 dirty"
		} else {
			summary = "0 total, 0 dirty"
		}
		blocks = append(blocks, streamingProjectStatusBlockExact(t, repo, "clean", remote, summary))
	}
	assert.Output(t, resp.Stdout, projectsStdoutV2(t, blocks...))

	probe := runProjectsStreamProbe(t, req)
	if probe.ExitCode != 0 {
		t.Fatalf("streaming probe exit %d", probe.ExitCode)
	}
	assertProjectsStreamsIncrementally(t, probe, req.MainRepo)
	t.Logf("streaming probe: first_byte_ms=%d total_ms=%d gap_ms=%d",
		probe.FirstByteMS, probe.TotalMS, probe.TotalMS-probe.FirstByteMS)
}
```