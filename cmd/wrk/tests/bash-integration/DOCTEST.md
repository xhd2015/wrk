# wrk bash integration Test Cases

## Version

**Layer: L2 in-process CLI** via `wrkcli.RunCLI` (runWrkOnce).
0.0.2

Decision tree covering `wrk --bash-integration`: print completion script, install/uninstall
lifecycle (dual bash profiles), read-only status, hidden `--complete` callback, and mutual
exclusion with normal wrk commands.

# DSN (Domain Specific Notion)

- **wrk CLI** — session-built binary; bash-integration modes are top-level flags mutually
  exclusive with create/list/done/etc.; they skip `events.jsonl` append.
- **WRK_HOME** — storage root (default `~/.wrk`); tests isolate per leaf at `{WorkRoot}/.wrk`
  unless overridden.
- **Fake HOME** — temp directory holding `~/.bash_profile` and `~/.bashrc` for install tests;
  profile markers append to both files.
- **integration/bash.sh** — completion script at `{WRK_HOME}/integration/bash.sh`; written on
  install; preserved on uninstall.
- **Profile marker block** — identical wrk marker in both profiles; sources bash.sh via
  `$WRK_HOME` when set.
- **projects.json** — basename completion reads unique sorted `filepath.Base(path)` entries;
  prefix-filtered candidates returned one per stdout line.
- **--complete callback** — `wrk --bash-integration --complete -- <words...> <cword>`; hidden
  from main help; drives bash `complete -o default -F _wrk wrk`.
- **path-like cur** — current word starting with `/`, `./`, or `../`. Go `Complete` returns no
  candidates (empty stdout, exit 0) so custom basenames/flags are not invented. Bash `_wrk`
  yields filename completion via `compopt -o default` and compspec `-o default`.

## Tree Overview

```
bash-integration/
├── print-script/basic/              # script: path-like yield + complete -o default
├── install/
│   ├── fresh/                       # writes bash.sh + dual profile markers
│   ├── idempotent/                  # pre-seeded state → no duplicate markers
│   └── dry-run/
│       ├── fresh/                   # preview only, no writes
│       └── already-installed/       # already installed message
├── uninstall/
│   ├── fresh/                       # strips both profile markers; keeps bash.sh
│   └── dry-run/
│       ├── with-marker/             # preview marker removal
│       └── already-uninstalled/     # already uninstalled message
├── status/
│   ├── installed/                   # script + both markers present
│   ├── not-installed/               # nothing present
│   └── partial/
│       ├── script-only/             # script without markers
│       └── marker-only/             # markers without script
├── wrk-home-override/custom/        # WRK_HOME=/custom install path + marker text
├── complete/
│   ├── basenames/                   # prefix filter on project basenames
│   ├── flags/                       # wrk -<tab> flag candidates
│   ├── list-dir/                    # wrk -l <tab>
│   ├── dep/                         # wrk --dep <tab>
│   ├── where/                       # wrk --where <tab>
│   ├── status/                      # wrk --status <tab>
│   ├── add-rm/                      # wrk --add/--rm <tab>
│   ├── empty-projects/              # no basenames; flags still work
│   └── path-like/                   # path-like cur → empty --complete stdout
│       ├── after-done/              # wrk --done <path-like>
│       │   ├── relative-dot/        # ./ex
│       │   ├── parent-relative/     # ../foo
│       │   └── absolute/            # /tmp/x
│       └── general/                 # positional path-like (./ ../ /)
└── mutual-exclusion/with-list/      # --bash-integration --list errors
```

## How to Run

```bash
doctest vet ./go-pkgs/cmd/wrk/tests/bash-integration
doctest test ./go-pkgs/cmd/wrk/tests/bash-integration
```

```go
import (
	"github.com/xhd2015/wrk/wrkcli"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"sync"
)

type CompleteCase struct {
	Name  string
	Words []string
	CWord int
}

type Request struct {
	WorkRoot string
	WrkHome  string
	FakeHome string
	RepoDir  string

	Mode     string // print | install | uninstall | status | complete | mutual
	DryRun   bool
	RunTwice bool
	PreInstall bool

	CLIArgs []string // when set, passed verbatim to wrk

	CompleteWords []string
	CompleteCWord int
	CompleteCases []CompleteCase

	ProjectPaths []string

	PreExistingBashProfile string
	PreExistingBashRC      string
	PreExistingBashSh      string
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string

	Home              string
	WrkHome           string
	BashShPath        string
	BashShContent     string
	BashProfilePath   string
	BashProfileContent string
	BashRCPath        string
	BashRCContent     string
	BashProfileMarkerCount int
	BashRCMarkerCount      int
	EventsPath        string
	EventsContent     string

	CompleteStdout map[string]string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.WorkRoot == "" || req.WrkHome == "" || req.FakeHome == "" {
		return nil, fmt.Errorf("Setup must initialize WorkRoot, WrkHome, and FakeHome")
	}

	if err := seedPreExistingState(req); err != nil {
		return nil, err
	}
	if len(req.ProjectPaths) > 0 {
		if err := writeProjectsJSON(req.WrkHome, req.ProjectPaths); err != nil {
			return nil, err
		}
	}

	bin := getWrkBin(t)

	if req.PreInstall {
		if _, _, _, err := runWrkOnce(t, req, bin, buildBashIntegrationArgs("install", false)); err != nil {
			return nil, fmt.Errorf("pre-install: %w", err)
		}
	}

	resp := &Response{
		Home:           req.FakeHome,
		WrkHome:        req.WrkHome,
		BashShPath:     bashShPath(req.WrkHome),
		BashProfilePath: filepath.Join(req.FakeHome, ".bash_profile"),
		BashRCPath:     filepath.Join(req.FakeHome, ".bashrc"),
		EventsPath:     filepath.Join(req.WrkHome, "events.jsonl"),
		CompleteStdout: make(map[string]string),
	}

	if len(req.CompleteCases) > 0 {
		for _, cc := range req.CompleteCases {
			stdout, stderr, code, err := runWrkOnce(t, req, bin, buildCompleteArgs(cc.Words, cc.CWord))
			if err != nil {
				return nil, fmt.Errorf("complete %s: %w", cc.Name, err)
			}
			resp.CompleteStdout[cc.Name] = stdout
			if cc.Name == req.CompleteCases[0].Name {
				resp.Stdout = stdout
				resp.Stderr = stderr
				resp.ExitCode = code
			}
		}
		captureFilesystemState(resp)
		return resp, nil
	}

	args := req.CLIArgs
	if args == nil {
		args = buildArgsFromMode(req)
	}

	stdout, stderr, code, err := runWrkOnce(t, req, bin, args)
	if err != nil {
		return nil, err
	}
	resp.Stdout = stdout
	resp.Stderr = stderr
	resp.ExitCode = code

	if req.RunTwice {
		stdout2, _, code2, err := runWrkOnce(t, req, bin, args)
		if err != nil {
			return nil, err
		}
		resp.Stdout = stdout2
		resp.ExitCode = code2
	}

	captureFilesystemState(resp)
	return resp, nil
}

func buildArgsFromMode(req *Request) []string {
	switch req.Mode {
	case "print":
		return []string{"--bash-integration"}
	case "install":
		args := []string{"--bash-integration", "--install"}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		return args
	case "uninstall":
		args := []string{"--bash-integration", "--uninstall"}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		return args
	case "status":
		return []string{"--bash-integration", "--status"}
	case "complete":
		return buildCompleteArgs(req.CompleteWords, req.CompleteCWord)
	case "mutual":
		return []string{"--bash-integration", "--list"}
	default:
		return []string{"--bash-integration"}
	}
}

func buildCompleteArgs(words []string, cword int) []string {
	args := []string{"--bash-integration", "--complete", "--"}
	args = append(args, words...)
	args = append(args, fmt.Sprintf("%d", cword))
	return args
}

func buildBashIntegrationArgs(action string, dryRun bool) []string {
	args := []string{"--bash-integration", "--" + action}
	if dryRun {
		args = append(args, "--dry-run")
	}
	return args
}

// bashNeedsHomeIsolation is true when the action reads/writes profile files under
// $HOME (install/uninstall/status). Those stay L3 via cmd.Env HOME=FakeHome.
// print-script, complete, and mutual-exclusion only need WRK_HOME + writers → L2.
func bashNeedsHomeIsolation(args []string) bool {
	for _, a := range args {
		switch a {
		case "--install", "--uninstall", "--status":
			return true
		}
	}
	return false
}

func runWrkOnce(t *testing.T, req *Request, bin string, args []string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	if !bashNeedsHomeIsolation(args) {
		var outBuf, errBuf bytes.Buffer
		code := wrkcli.RunCLI(args, wrkcli.RunOptions{
			Stdout:  &outBuf,
			Stderr:  &errBuf,
			Dir:     req.RepoDir,
			WrkHome: req.WrkHome,
		})
		return outBuf.String(), errBuf.String(), code, nil
	}
	if bin == "" {
		bin = getWrkBin(t)
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = []string{
		"HOME=" + req.FakeHome,
		"WRK_HOME=" + req.WrkHome,
		"PATH=" + os.Getenv("PATH"),
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	exitCode = 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return "", "", 0, runErr
		}
	}
	return outBuf.String(), errBuf.String(), exitCode, nil
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
	resp.EventsContent, _ = readFileIfExists(resp.EventsPath)
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