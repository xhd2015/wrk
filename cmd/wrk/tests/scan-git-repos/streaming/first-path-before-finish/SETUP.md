# Scenario

**Feature**: first main path appears on stdout before --scan-git-repos finishes (later root still scanning)

```
root-first/main-first + root-later/[pads…]/zzz-main
  -> wrk --scan-git-repos --no-cache root-first root-later
  -> stdout first line: abs(main-first) while pads/zzz-main still walking
  -> process eventually exits 0 with both mains on stdout + projects.json
  -> probe: total_ms - first_byte_ms >= 40ms (incremental, not batch-at-end)
```

## Preconditions

- Two CLI roots so discovery prints first main, then continues on a slow second root.
- `root-later` has many empty pad dirs (`a00000`…) and main `zzz-main` so walk
  of the second root continues after the first stdout path line (measurable gap).
- `--no-cache` keeps the walk from short-circuiting via a warm cache.
- Cwd remains non-git `{WorkRoot}`.
- Root `Run` may complete a full scan; Assert resets `projects.json` and re-runs
  via `runScanGitReposStreamProbe` for first-byte / total timing.

## Steps

1. Create `root-first/main-first` as a main git repo.
2. Create `root-later` with padding dirs and `zzz-main` main git repo.
3. Set Args: `--scan-git-repos --no-cache <root-first> <root-later>`.
4. Assert runs stream timing probe (fresh projects.json, pipe, firstByteMS vs totalMS).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	ensureScanStreamProbeHelpersUsed()

	rootFirst := filepath.Join(req.WorkRoot, "root-first")
	rootLater := filepath.Join(req.WorkRoot, "root-later")
	mkdirAll(t, rootFirst)
	mkdirAll(t, rootLater)

	mainFirst := initScanMainRepo(t, rootFirst, "main-first")
	// Pads before zzz-main keep the second root mid-walk after first path is printed.
	seedScanPaddingDirs(t, rootLater, 2000)
	mainLater := initScanMainRepo(t, rootLater, "zzz-main")

	req.MainRepo = mainFirst
	req.SecondRepo = mainLater
	req.Args = []string{"--scan-git-repos", "--no-cache", rootFirst, rootLater}
	return nil
}
```
