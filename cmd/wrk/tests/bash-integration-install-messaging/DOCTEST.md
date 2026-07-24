# wrk bash-integration install messaging

## Version

**Layer: L2 in-process CLI** via `wrkcli.RunCLI` (runWrkOnce).
0.0.2

Decision tree for clearer per-component and summary stdout from
`wrk --bash-integration --install` and matching `--dry-run`. Covers script and
profile-marker status vocabulary (`installed` / `updated` / `is up to date` and
dry-run `would install` / `would update` / `is up to date`).

# DSN (Domain Specific Notion)

- **wrk CLI** — session-built binary; `--bash-integration --install` writes
  integration assets under isolated `WRK_HOME` and fake `HOME`; skips
  `events.jsonl`.
- **WRK_HOME** — storage root for `integration/bash.sh`; tests isolate per leaf
  at `{WorkRoot}/.wrk`.
- **Fake HOME** — temp directory holding `~/.bash_profile` and `~/.bashrc`.
- **integration/bash.sh** — completion + wrapper script; **installed** when
  missing, **updated** when present but not byte-equal to the embedded script,
  **is up to date** when identical.
- **Profile markers** — identical wrk marker blocks in `.bash_profile` and
  `.bashrc`; only appended once (**installed** vs **is up to date**); never
  rewritten (no marker `updated`).
- **Install report** — four stdout lines (summary + script + two markers) plus
  trailing blank line; exit 0 for all successful summary outcomes.
- **Dry-run report** — same four-line shape with `would install` /
  `would update` vocabulary for planned writes; no filesystem writes; must
  detect outdated script content (not mere presence).

## Tree Overview

```
bash-integration-install-messaging/
├── install/                         # real --install (writes)
│   ├── fresh/                       # nothing → installed
│   ├── outdated-script/             # old bash.sh + markers → updated
│   ├── current/                     # all current → is up to date
│   ├── markers-missing/             # current script, no markers → updated
│   └── one-marker-missing/          # current script, one marker → updated
└── dry-run/                         # --install --dry-run (no writes)
    ├── fresh/                       # would install
    ├── outdated-script/             # would update (outdated content)
    ├── current/                     # is up to date
    └── markers-missing/             # would update (markers only)
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | install/fresh | summary `installed`; script+both markers `(installed)` |
| 2 | install/outdated-script | summary `updated`; script `(updated)`; markers up to date |
| 3 | install/current | summary `is up to date`; all components up to date |
| 4 | install/markers-missing | summary `updated`; script up to date; both markers installed |
| 5 | install/one-marker-missing | summary `updated`; one marker installed, one up to date |
| 6 | dry-run/fresh | summary `would install`; would-install components; no writes |
| 7 | dry-run/outdated-script | summary `would update`; script would update; no writes |
| 8 | dry-run/current | summary `is up to date`; no writes |
| 9 | dry-run/markers-missing | summary `would update`; markers would install; no writes |

## Compatibility notes

- Existing `bash-integration/install/fresh` and `idempotent` assert files only
  (not empty stdout) — they should stay green when install **adds** the report.
- Existing `bash-integration/install/dry-run/*` assert the **old** dry-run
  vocabulary (`dry-run: would write…`, `already installed`). Implementer must
  update those leaves to the new four-line contract (or delete them once this
  tree owns dry-run messaging).
- `followup-cd/script-surface/install/*` asserts script content only — stays green.

## How to Run

```sh
doctest vet ./go-pkgs/cmd/wrk/tests/bash-integration-install-messaging
doctest test ./go-pkgs/cmd/wrk/tests/bash-integration-install-messaging
doctest test ./go-pkgs/cmd/wrk/tests/bash-integration-install-messaging/install/fresh
doctest test ./go-pkgs/cmd/wrk/tests/bash-integration-install-messaging/dry-run/outdated-script
```

```go
import (
	"github.com/xhd2015/wrk/wrkcli"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"sync"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	FakeHome string
	RepoDir  string

	Mode       string // install
	DryRun     bool
	PreInstall bool

	PreExistingBashProfile string
	PreExistingBashRC      string
	PreExistingBashSh      string

	// SeedCurrentScript writes the live embedded bash.sh (from
	// `wrk --bash-integration` print) before the tested run, without markers
	// unless PreExistingBashProfile/RC are also set.
	SeedCurrentScript bool
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string

	Home               string
	WrkHome            string
	BashShPath         string
	BashShContent      string
	BashProfilePath    string
	BashProfileContent string
	BashRCPath         string
	BashRCContent      string
	BashProfileMarkerCount int
	BashRCMarkerCount      int
	EventsPath         string

	// Pre-run snapshots for dry-run "no writes" checks.
	BeforeBashShContent      string
	BeforeBashShExists       bool
	BeforeBashProfileContent string
	BeforeBashProfileExists  bool
	BeforeBashRCContent      string
	BeforeBashRCExists       bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.WorkRoot == "" || req.WrkHome == "" || req.FakeHome == "" {
		return nil, fmt.Errorf("Setup must initialize WorkRoot, WrkHome, and FakeHome")
	}

	bin := getWrkBin(t)

	if req.SeedCurrentScript {
		script, err := captureEmbeddedBashSh(t, req, bin)
		if err != nil {
			return nil, fmt.Errorf("capture embedded bash.sh: %w", err)
		}
		req.PreExistingBashSh = script
	}

	if err := seedPreExistingState(req); err != nil {
		return nil, err
	}

	if req.PreInstall {
		if _, _, _, err := runWrkOnce(t, req, bin, []string{"--bash-integration", "--install"}); err != nil {
			return nil, fmt.Errorf("pre-install: %w", err)
		}
	}

	resp := &Response{
		Home:            req.FakeHome,
		WrkHome:         req.WrkHome,
		BashShPath:      bashShPath(req.WrkHome),
		BashProfilePath: filepath.Join(req.FakeHome, ".bash_profile"),
		BashRCPath:      filepath.Join(req.FakeHome, ".bashrc"),
		EventsPath:      filepath.Join(req.WrkHome, "events.jsonl"),
	}

	resp.BeforeBashShContent, resp.BeforeBashShExists = readFileIfExists(resp.BashShPath)
	resp.BeforeBashProfileContent, resp.BeforeBashProfileExists = readFileIfExists(resp.BashProfilePath)
	resp.BeforeBashRCContent, resp.BeforeBashRCExists = readFileIfExists(resp.BashRCPath)

	args := buildArgs(req)
	stdout, stderr, code, err := runWrkOnce(t, req, bin, args)
	if err != nil {
		return nil, err
	}
	resp.Stdout = stdout
	resp.Stderr = stderr
	resp.ExitCode = code

	captureFilesystemState(resp)
	return resp, nil
}

func buildArgs(req *Request) []string {
	args := []string{"--bash-integration", "--install"}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	return args
}

func runWrkOnce(t *testing.T, req *Request, bin string, args []string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	_ = bin
	var outBuf, errBuf bytes.Buffer
	code := wrkcli.RunCLI(args, wrkcli.RunOptions{
		Stdout:  &outBuf,
		Stderr:  &errBuf,
		Dir:     req.RepoDir,
		WrkHome: req.WrkHome,
		Home:    req.FakeHome,
	})
	return outBuf.String(), errBuf.String(), code, nil
}


func captureEmbeddedBashSh(t *testing.T, req *Request, bin string) (string, error) {
	t.Helper()
	stdout, stderr, code, err := runWrkOnce(t, req, bin, []string{"--bash-integration"})
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("print-script exit %d; stderr=%s", code, stderr)
	}
	if stdout == "" {
		return "", fmt.Errorf("print-script returned empty stdout")
	}
	return stdout, nil
}

func findModuleRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Process-local wrk binary (one-process suite; in-memory mutex, not session.Once/flock).
var (
	wrkBinMu   sync.Mutex
	wrkBinPath string
	wrkBinErr  error
	// wrkModRoot set once via noteModRoot (sync.Once); not per-leaf Setup writes.
	wrkModOnce sync.Once
	wrkModRoot string
)


// noteModRoot records module root once per process (sync.Once). Safe under t.Parallel.
// Prefer this over writing wrkModRoot from every leaf Setup.
func noteModRoot(d *session.Doctest) {
	if d == nil {
		return
	}
	wrkModOnce.Do(func() {
		wrkModRoot = findModuleRoot(d.DOCTEST_ROOT)
	})
}

func getWrkBin(t *testing.T) string {
	t.Helper()
	wrkBinMu.Lock()
	defer wrkBinMu.Unlock()
	if wrkBinPath != "" || wrkBinErr != nil {
		if wrkBinErr != nil {
			t.Fatal(wrkBinErr)
		}
		return wrkBinPath
	}
	if wrkModRoot == "" {
		t.Fatal("wrkModRoot unset; root Setup must call noteModRoot first")
	}
	dir, err := os.MkdirTemp("", "wrk-doctest-bin-")
	if err != nil {
		wrkBinErr = err
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "wrk")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/wrk")
	cmd.Dir = wrkModRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		wrkBinErr = fmt.Errorf("build wrk: %v\n%s", err, out)
		t.Fatal(wrkBinErr)
	}
	wrkBinPath = bin
	return bin
}

func captureFilesystemState(resp *Response) {
	resp.BashShContent, _ = readFileIfExists(resp.BashShPath)
	resp.BashProfileContent, _ = readFileIfExists(resp.BashProfilePath)
	resp.BashRCContent, _ = readFileIfExists(resp.BashRCPath)
	resp.BashProfileMarkerCount = countWrkMarkers(resp.BashProfileContent)
	resp.BashRCMarkerCount = countWrkMarkers(resp.BashRCContent)
}

func seedPreExistingState(req *Request) error {
	if req.PreExistingBashSh != "" {
		if err := os.MkdirAll(filepath.Dir(bashShPath(req.WrkHome)), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(bashShPath(req.WrkHome), []byte(req.PreExistingBashSh), 0o644); err != nil {
			return err
		}
	}
	if req.PreExistingBashProfile != "" {
		if err := os.WriteFile(filepath.Join(req.FakeHome, ".bash_profile"), []byte(req.PreExistingBashProfile), 0o644); err != nil {
			return err
		}
	}
	if req.PreExistingBashRC != "" {
		if err := os.WriteFile(filepath.Join(req.FakeHome, ".bashrc"), []byte(req.PreExistingBashRC), 0o644); err != nil {
			return err
		}
	}
	return nil
}
```
