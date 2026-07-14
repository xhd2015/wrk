# Scenario

**Feature**: wrk --projects streams stdout incrementally (per-project blocks as they complete)

```
fast project (aaa, no worktrees) + slow project (zzz, 12 linked worktrees)
-> wrk --projects -> first stdout bytes arrive before run_end (fast block while slow still gathering)
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- Lexicographic order prints `aaa` before `zzz`; streaming means the `aaa` block appears while `zzz` is still gathering.

## Context

`runProjects` today gathers every project into `results[]` (`wg.Wait()`), then prints all blocks. That buffers stdout until the slowest project finishes — bad UX when many projects or heavy worktree scans are registered.

Streaming probe helpers run `wrk --projects` with a stdout pipe, record `firstByteMS` / `totalMS`, and capture the first read chunk prefix.

```go
import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/gitops/git"
	"time"
)

type projectsStreamProbe struct {
	FirstByteMS  int64
	TotalMS      int64
	FirstChunk   string
	FullStdout   string
	ExitCode     int
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return nil
}

func runProjectsStreamProbe(t *testing.T, req *Request) projectsStreamProbe {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, "--projects")
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)

	stdoutR, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start wrk --projects: %v", err)
	}

	type readResult struct {
		firstByteMS int64
		firstChunk  string
		fullStdout  string
		readErr     error
	}
	readDone := make(chan readResult, 1)
	go func() {
		var firstByteMS int64 = -1
		var firstChunk, rest bytes.Buffer
		buf := make([]byte, 4096)
		for {
			n, readErr := stdoutR.Read(buf)
			if n > 0 {
				if firstByteMS < 0 {
					firstByteMS = time.Since(start).Milliseconds()
					firstChunk.Write(buf[:n])
				} else {
					rest.Write(buf[:n])
				}
			}
			if readErr == io.EOF {
				readDone <- readResult{
					firstByteMS: firstByteMS,
					firstChunk:  firstChunk.String(),
					fullStdout:  firstChunk.String() + rest.String(),
				}
				return
			}
			if readErr != nil {
				readDone <- readResult{readErr: readErr}
				return
			}
		}
	}()

	waitErr := cmd.Wait()
	totalMS := time.Since(start).Milliseconds()

	var rr readResult
	select {
	case rr = <-readDone:
	case <-time.After(30 * time.Second):
		t.Fatal("timeout reading wrk --projects stdout (30s)")
	}
	if rr.readErr != nil {
		t.Fatalf("read stdout: %v", rr.readErr)
	}

	exitCode := 0
	if waitErr != nil {
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("wait wrk --projects: %v", waitErr)
		}
	}

	return projectsStreamProbe{
		FirstByteMS: rr.firstByteMS,
		TotalMS:     totalMS,
		FirstChunk:  rr.firstChunk,
		FullStdout:  rr.fullStdout,
		ExitCode:    exitCode,
	}
}

func assertProjectsStreamsIncrementally(t *testing.T, probe projectsStreamProbe, fastRepoPath string) {
	t.Helper()
	const minTotalMS = int64(80)
	const minLeadMS = int64(40)

	if probe.FirstByteMS < 0 {
		t.Fatalf("no stdout until process exit (buffered); total_ms=%d", probe.TotalMS)
	}
	if probe.TotalMS < minTotalMS {
		t.Fatalf("total_ms=%d too short for streaming probe (want >=%d)", probe.TotalMS, minTotalMS)
	}
	gap := probe.TotalMS - probe.FirstByteMS
	if gap < minLeadMS {
		t.Fatalf("stdout not incremental: first_byte_ms=%d total_ms=%d gap_ms=%d (want gap >= %d; gather-then-print buffers until slowest project finishes)",
			probe.FirstByteMS, probe.TotalMS, gap, minLeadMS)
	}

	fastDir := "Dir:          " + resolvePath(t, fastRepoPath)
	if !strings.HasPrefix(probe.FirstChunk, fastDir) {
		t.Fatalf("first stdout chunk should start with fast project block %q, got:\n%q", fastDir, probe.FirstChunk)
	}
}

func initStreamingStatusRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func setupStreamingBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func setupStreamingTrackedMainRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initStreamingStatusRepo(t, repo, subject)
	runGitIsolated(t, repo, "remote", "add", "origin", originBare)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
	return repo
}

func addStreamingLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func recordStreamingProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
}

func setupFastNoUpstreamRepo(t *testing.T, workRoot, name string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initStreamingStatusRepo(t, repo, name+" fast")
	return repo
}

func setupSlowManyWorktreesRepo(t *testing.T, req *Request, name string, worktreeCount int) string {
	t.Helper()
	origin := setupStreamingBareOrigin(t, req.WorkRoot, name+"-origin")
	repo := setupStreamingTrackedMainRepo(t, req.WorkRoot, name, origin, name+" slow")
	for i := 1; i <= worktreeCount; i++ {
		branch := fmt.Sprintf("wt-%d", i)
		addStreamingLinkedWorktree(t, repo, branch, branch)
	}
	return repo
}

func streamingStatusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	branch := gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	return "Branch:       " + branch
}

func streamingStatusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func streamingProjectDirLine(t *testing.T, mainRepo string) string {
	t.Helper()
	return "Dir:          " + resolvePath(t, mainRepo)
}

func streamingRemoteBriefFromResult(result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs merge back(+%d %s)", result.CommitsAheadB, commitWord)
	case git.BranchRelationBIsAncestorOfA:
		return "needs pull"
	case git.BranchRelationDiverged:
		diverged := result.CommitsAheadA + result.CommitsAheadB
		commitWord := "commit"
		if diverged != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("diverged(%d %s)", diverged, commitWord)
	default:
		return fmt.Sprintf("unknown branch relation %v", result.Relation)
	}
}

func streamingCompareWithRemoteField(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
	t.Helper()
	if upstreamRef == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstreamRef, currentBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, upstreamRef, currentBranch, err)
	}
	return "Remote:       " + streamingRemoteBriefFromResult(result)
}

func streamingProjectStatusBlockExact(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string) string {
	t.Helper()
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s\nWorktrees:    %s",
		streamingProjectDirLine(t, mainRepo),
		streamingStatusBranchLine(t, mainRepo),
		streamingStatusCommitLine(t, mainRepo),
		statusLine,
		compareRemoteField,
		worktreesSummary,
	)
}

func streamingProjectsBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	got := strings.Count(stdout, "Dir:          ")
	if got != wantBlocks {
		t.Fatalf("expected %d project blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between project blocks, got:\n%s", stdout)
	}
}

func ensureOutputStreamingHelpersUsed() {
	_ = runProjectsStreamProbe
	_ = assertProjectsStreamsIncrementally
	_ = setupFastNoUpstreamRepo
	_ = setupSlowManyWorktreesRepo
	_ = recordStreamingProject
	_ = initStreamingStatusRepo
	_ = setupStreamingBareOrigin
	_ = setupStreamingTrackedMainRepo
	_ = addStreamingLinkedWorktree
	_ = streamingStatusBranchLine
	_ = streamingStatusCommitLine
	_ = streamingProjectDirLine
	_ = streamingRemoteBriefFromResult
	_ = streamingCompareWithRemoteField
	_ = streamingProjectStatusBlockExact
	_ = streamingProjectsBlocksSeparated
}
```