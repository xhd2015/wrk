package wrkcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	xgocmd "github.com/xhd2015/xgo/support/cmd"
)

var invocationVerbose bool

func setInvocationVerbose(v bool) {
	invocationVerbose = v
}

func logGitCommand(args []string) {
	if !invocationVerbose || !isMajorGitCommand(args) {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(os.Stderr, "[%s] $ git %s\n", ts, strings.Join(args, " "))
}

// logGoCommand logs a major go invocation under -v in the same timestamp style
// as logGitCommand, e.g. `[YYYY-MM-DD HH:MM:SS] $ go -C <dir> mod tidy`.
func logGoCommand(args []string) {
	if !invocationVerbose {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(os.Stderr, "[%s] $ go %s\n", ts, strings.Join(args, " "))
}

// goModTidy runs `go mod tidy` in dir. Under -v: timestamped pre-line on stderr
// then streams child stdout+stderr to process stderr (so wrk stdout stays clean).
// Without -v: silent success (no pre-line). Always captures child streams so
// failure wraps include the go diagnostic body, not only exit status 1.
func goModTidy(dir string) error {
	absDir := dir
	if abs, err := filepath.Abs(dir); err == nil {
		absDir = abs
	}
	if invocationVerbose {
		logGoCommand([]string{"-C", absDir, "mod", "tidy"})
	}
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = absDir
	var buf bytes.Buffer
	if invocationVerbose {
		// Stream and capture so path-only stdout stays clean and failures
		// still carry the go diagnostic body.
		mw := io.MultiWriter(os.Stderr, &buf)
		cmd.Stdout = mw
		cmd.Stderr = mw
	} else {
		cmd.Stdout = &buf
		cmd.Stderr = &buf
	}
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(buf.String())
		if msg != "" {
			return fmt.Errorf("failed to execute go mod tidy: %w\n%s", err, msg)
		}
		return fmt.Errorf("failed to execute go mod tidy: %w", err)
	}
	return nil
}

type goModRequire struct {
	path, version string
}

func listGoModRequires(dir string) ([]goModRequire, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Require []struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"Require"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, err
	}
	var reqs []goModRequire
	for _, r := range parsed.Require {
		if r.Path == "" {
			continue
		}
		reqs = append(reqs, goModRequire{path: r.Path, version: r.Version})
	}
	return reqs, nil
}

func isMajorGitCommand(args []string) bool {
	i := 0
	for i < len(args) && args[i] == "-C" {
		i += 2
	}
	if i >= len(args) {
		return false
	}
	switch args[i] {
	case "fetch", "checkout", "merge", "rebase", "stash":
		return true
	case "worktree":
		if i+1 < len(args) {
			switch args[i+1] {
			case "add", "remove", "move":
				return true
			}
		}
	case "branch":
		if i+1 < len(args) {
			switch args[i+1] {
			case "-D", "-m", "-b":
				return true
			}
		}
	}
	return false
}

// gitCommand returns *exec.Cmd for callers that need custom Stdout/Stderr
// streaming or .Run() control (e.g. runGitWorktreeAdd). Prefer gitRun /
// gitCombinedOutputHelpers that use xgo/support/cmd for non-interactive capture.
// os/exec is kept here intentionally for that streaming control surface.
func gitCommand(args ...string) *exec.Cmd {
	logGitCommand(args)
	return exec.Command("git", args...)
}

// gitCommandDir builds a git command with Dir set. See gitCommand for os/exec rationale.
func gitCommandDir(repoPath string, args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd
}

// gitCommandWithEnv builds a git command with extra env. See gitCommand for os/exec rationale.
func gitCommandWithEnv(repoPath string, extraEnv []string, args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd
}

// gitRunDir runs git in repoPath via xgo/support/cmd (non-interactive).
func gitRunDir(repoPath string, args ...string) error {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	return xgocmd.Dir(repoPath).Run("git", args...)
}

// gitOutputDir captures git stdout via xgo/support/cmd (non-interactive).
// Note: git stderr may still inherit the process stderr (e.g. "fatal: no upstream").
// For TUI soft probes, use gitOutputDirCapture and route stderr into the Log panel.
func gitOutputDir(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	return xgocmd.Dir(repoPath).Output("git", args...)
}

// gitOutputDirCapture captures git stdout and stderr without inheriting the tty.
// Stderr must be treated as normal log lines by the caller (never discarded silently,
// never printed to the real terminal while the TUI owns the screen).
// Does not call logGitCommand (verbose mode would write to process stderr).
func gitOutputDirCapture(repoPath string, args ...string) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	err = xgocmd.Dir(repoPath).Stdout(&outBuf).Stderr(&errBuf).Run("git", args...)
	return outBuf.String(), errBuf.String(), err
}

// splitCapturedLogLines splits captured git stderr/stdout into non-empty log lines.
func splitCapturedLogLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// gitCombinedRunDir runs git capturing combined stdout+stderr (error messages).
func gitCombinedRunDir(repoPath string, extraEnv []string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	var buf bytes.Buffer
	builder := xgocmd.Dir(repoPath).Stdout(&buf).Stderr(&buf)
	if len(extraEnv) > 0 {
		builder = builder.Env(extraEnv)
	}
	err := builder.Run("git", args...)
	return buf.Bytes(), err
}

// runGitWorktreeAdd runs git worktree add. When verbose, streams stdout+stderr to
// os.Stderr so git's own progress lines appear alongside the pre-command log.
// Streaming requires *exec.Cmd; non-verbose path uses CombinedOutput on the cmd.
func runGitWorktreeAdd(cmd *exec.Cmd) error {
	if invocationVerbose {
		// Interactive/streaming: keep os/exec Stdout/Stderr wiring.
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git worktree add: %w", err)
		}
		return nil
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, out)
	}
	return nil
}
