# Scenario

**Feature**: wrk --fetch opt-in upstream refresh and -v verbose git command logging

```
# --fetch valid only with --projects or --status; default skips git fetch
wrk --projects [--fetch] -> Remote: from local upstream tracking refs (or fetch first)

# -v logs major git subprocesses to stderr without changing stdout
wrk <mode> -v -> stderr [timestamp] $ git <args...>
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk` and `WRK_DATE=2026-06-30`.

## Context

- Upstream fixture helpers mirror `projects/remote-brief/` patterns.
- Verbose stderr assertions use `<contains>` for git subcommand substrings.
- Status `Remote:` helpers build root-block templates with field order: Dir, Branch, Commit, Status, Remote.

```go
import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/gitops/git"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

func initFetchVerboseRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func setupFetchVerboseBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func setupFetchVerboseTrackedRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initFetchVerboseRepo(t, repo, subject)
	runGitIsolated(t, repo, "remote", "add", "origin", originBare)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
	return repo
}

func pushCommitToFetchVerboseOrigin(t *testing.T, workRoot, originBare, filename, content, subject string) {
	t.Helper()
	cloneDir := filepath.Join(workRoot, "origin-push-clone")
	runGitIsolated(t, workRoot, "clone", originBare, cloneDir)
	writeFile(t, filepath.Join(cloneDir, filename), content)
	runGitIsolated(t, cloneDir, "add", filename)
	runGitIsolated(t, cloneDir, "commit", "-m", subject)
	runGitIsolated(t, cloneDir, "push", "origin", "main")
}

func fetchInMainRepo(t *testing.T, mainRepo string) {
	t.Helper()
	runGitIsolated(t, mainRepo, "fetch", "origin")
}

func recordFetchVerboseProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
}

func remoteBriefFromResult(result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs push(+%d %s)", result.CommitsAheadB, commitWord)
	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs pull(%d %s behind)", result.CommitsAheadA, commitWord)
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

func remoteFieldLine(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
	t.Helper()
	if upstreamRef == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstreamRef, currentBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, upstreamRef, currentBranch, err)
	}
	return "Remote:       " + remoteBriefFromResult(result)
}

func statusRootBlockWithRemotePlain(t *testing.T, mainRepo, statusLine, remoteLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       %s\n%s",
		statusBranchLine(t, mainRepo),
		statusCommitLine(t, mainRepo),
		statusLine,
		remoteLine,
	)
}

func statusRootBlockWithRemoteTemplate(t *testing.T, mainRepo, statusLine, remoteLine string) string {
	t.Helper()
	return v2StdoutTemplate(statusRootBlockWithRemotePlain(t, mainRepo, statusLine, remoteLine))
}

func projectsRemoteBlockTemplate(t *testing.T, mainRepo, statusLine, remoteLine, worktreesSummary string) string {
	t.Helper()
	block := fmt.Sprintf("Dir:          %s\nBranch:       %s\nCommit:       %s  %s\nStatus:       %s\n%s\nWorktrees:    %s",
		resolvePath(t, mainRepo),
		gitOutputIsolated(t, mainRepo, "rev-parse", "--abbrev-ref", "HEAD"),
		gitOutputIsolated(t, mainRepo, "rev-parse", "--short=7", "HEAD"),
		gitOutputIsolated(t, mainRepo, "log", "-1", "--pretty=%s"),
		statusLine,
		remoteLine,
		worktreesSummary,
	)
	return v2StdoutTemplate(block)
}

func statusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	return "Branch:       " + gitOutputIsolated(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

func statusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutputIsolated(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutputIsolated(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func addLinkedWorktreeInRepo(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func masterIdenticalField(t *testing.T, mainRepo, mainBranch, wtBranch string) string {
	t.Helper()
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, mainBranch, wtBranch, err)
	}
	if result.Relation != git.BranchRelationSame {
		t.Fatalf("expected identical Master relation, got %v", result.Relation)
	}
	return "Master:       identical"
}

func assertStderrContainsGitSubcommand(t *testing.T, stderr, subcommand string) {
	t.Helper()
	if !strings.Contains(stderr, "git ") || !strings.Contains(stderr, subcommand) {
		t.Fatalf("stderr should contain git %q log line, got %q", subcommand, stderr)
	}
}

func assertStderrNoGitSubcommand(t *testing.T, stderr, subcommand string) {
	t.Helper()
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "git ") && strings.Contains(line, subcommand) {
			t.Fatalf("stderr should not log git %q, got %q", subcommand, stderr)
		}
	}
}

func assertStderrVerboseFormat(t *testing.T, stderr string) {
	t.Helper()
	re := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \$ git `)
	found := false
	for _, line := range strings.Split(stderr, "\n") {
		if re.MatchString(line) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("stderr should contain verbose git log line matching timestamp format, got %q", stderr)
	}
}

func assertStderrContainsWorktreeAddOutput(t *testing.T, stderr string) {
	t.Helper()
	if strings.Contains(stderr, "Preparing worktree") || strings.Contains(stderr, "HEAD is now at") {
		return
	}
	t.Fatalf("stderr should contain git worktree add subprocess output (Preparing worktree or HEAD is now at), got %q", stderr)
}

func assertStdoutNoRemoteField(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "Remote:") {
		t.Fatalf("stdout should not contain Remote:, got:\n%s", stdout)
	}
}

func assertFetchInvalidModeStderr(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assertContains(t, resp.Stderr, "--fetch is only valid with --projects or --status")
}

func ensureFetchVerboseHelpersUsed() {
	_ = initFetchVerboseRepo
	_ = setupFetchVerboseBareOrigin
	_ = setupFetchVerboseTrackedRepo
	_ = pushCommitToFetchVerboseOrigin
	_ = fetchInMainRepo
	_ = recordFetchVerboseProject
	_ = remoteBriefFromResult
	_ = remoteFieldLine
	_ = statusRootBlockWithRemotePlain
	_ = statusRootBlockWithRemoteTemplate
	_ = projectsRemoteBlockTemplate
	_ = statusBranchLine
	_ = statusCommitLine
	_ = addLinkedWorktreeInRepo
	_ = masterIdenticalField
	_ = assertStderrContainsGitSubcommand
	_ = assertStderrNoGitSubcommand
	_ = assertStderrVerboseFormat
	_ = assertStderrContainsWorktreeAddOutput
	_ = assertStdoutNoRemoteField
	_ = assertFetchInvalidModeStderr
	_ = resolvePath
}
```