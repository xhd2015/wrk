# Scenario

**Feature**: SIGINT after first newly printed main → exit 130, warning, partial projects

```
root-first/main-first + root-later/[pads…]/zzz-main
  -> wrk --scan-git-repos --no-cache root-first root-later
  -> stdout first line: abs(main-first)
  -> harness SIGINT
  -> exit 130
  -> stderr warning: interrupted + progress saved
  -> projects.json: main-first source=scan; zzz-main not required if unvisited
```

## Preconditions

- Two CLI roots so discovery order is first main then later main.
- `root-later` has many empty pad dirs (`a00000`…) and main `zzz-main` so walk
  of the second root is still in progress when SIGINT is delivered after the
  first stdout path line (reduces flaky full-completion before signal).
- `--no-cache` keeps the walk from short-circuiting via a warm cache.
- Cwd remains non-git `{WorkRoot}`.

## Steps

1. Create `root-first/main-first` as a main git repo.
2. Create `root-later` with padding dirs and `zzz-main` main git repo.
3. Set Args: `--scan-git-repos --no-cache <root-first> <root-later>`.
4. Root `Run` may complete a full scan (ignored); Assert resets `projects.json`
   and re-runs via `runScanGitReposSIGINTAfterFirstStdout`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	ensureScanInterruptHelpersUsed()

	rootFirst := filepath.Join(req.WorkRoot, "root-first")
	rootLater := filepath.Join(req.WorkRoot, "root-later")
	mkdirAll(t, rootFirst)
	mkdirAll(t, rootLater)

	mainFirst := initScanMainRepo(t, rootFirst, "main-first")
	// Pads before zzz-main give the SIGINT harness a mid-walk window on root-later.
	seedScanPaddingDirs(t, rootLater, 2000)
	mainLater := initScanMainRepo(t, rootLater, "zzz-main")

	req.MainRepo = mainFirst
	req.SecondRepo = mainLater
	req.Args = []string{"--scan-git-repos", "--no-cache", rootFirst, rootLater}
	return nil
}
```
