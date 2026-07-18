# Scenario

**Feature**: wrk create fails when git post-checkout hooks require git-lfs but process PATH is stripped

```
# LFS-enabled repo + post-checkout hook requiring git-lfs on PATH
# git-lfs shim lives under $HOME/.local/bin (not on stripped PATH)
wrk -> git worktree add -> hook misses git-lfs -> exit 1 (expected)
```

## Preconditions

- Git must be available.
- Tests install a fake `git-lfs` executable under `{FakeHome}/.local/bin/`.
- `UseMinimalPath` runs wrk with `HOME={FakeHome}` and `PATH={WorkRoot}/minimal-bin`
  only (symlinks to `git`/`sh`/etc., **not** system `git-lfs`). Ubuntu runners ship
  `git-lfs` in `/usr/bin`, so `PATH=/usr/bin:/bin` is not a valid "stripped" PATH there.

## Steps

- Configure the source repo with `filter.lfs.required=true` and a `core.hookspath` post-checkout hook that errors when `git-lfs` is missing (mirrors real global hooks).
- Run `wrk` with stripped PATH; the hook fails because wrk does not augment PATH for git subprocesses.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func initFakeHomeWithGitLFS(t *testing.T, workRoot string) string {
	t.Helper()
	home := filepath.Join(workRoot, "fakehome")
	binDir := filepath.Join(home, ".local", "bin")
	mkdirAll(t, binDir)
	writeFile(t, filepath.Join(binDir, "git-lfs"), "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(filepath.Join(binDir, "git-lfs"), 0755); err != nil {
		t.Fatalf("chmod git-lfs: %v", err)
	}
	// Controlled PATH without system git-lfs (see ensureMinimalBinWithoutGitLFS).
	_ = ensureMinimalBinWithoutGitLFS(t, workRoot)
	return home
}

// ensureMinimalBinWithoutGitLFS builds WorkRoot/minimal-bin with symlinks to
// common tools needed by git/hooks, deliberately excluding git-lfs so Ubuntu CI
// (which installs git-lfs under /usr/bin) still exercises the missing-lfs path.
func ensureMinimalBinWithoutGitLFS(t *testing.T, workRoot string) string {
	t.Helper()
	dir := filepath.Join(workRoot, "minimal-bin")
	mkdirAll(t, dir)
	// Keep the set small: git + shell builtins/tools hooks and git may invoke.
	for _, name := range []string{
		"git", "sh", "bash", "rm", "cat", "sed", "tr", "uname",
		"dirname", "basename", "mkdir", "cp", "mv", "ls", "chmod",
		"true", "false", "echo", "printf", "env", "test", "[",
		"awk", "head", "tail", "grep", "sort", "diff", "mktemp",
		"touch", "date", "pwd", "which", "command",
	} {
		src, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if filepath.Base(src) == "git-lfs" {
			continue
		}
		dst := filepath.Join(dir, name)
		_ = os.Remove(dst)
		if err := os.Symlink(src, dst); err != nil {
			t.Fatalf("symlink %s -> %s: %v", src, dst, err)
		}
	}
	// Refuse to proceed if git-lfs still resolves via this PATH (misconfigured).
	pathEnv := dir + string(os.PathListSeparator) + dir
	cmd := exec.Command("sh", "-c", "command -v git-lfs")
	cmd.Env = []string{"PATH=" + dir}
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("minimal-bin still resolves git-lfs: %s", strings.TrimSpace(string(out)))
	}
	_ = pathEnv
	return dir
}

func initGitRepoWithLFSHooks(t *testing.T, repoDir, hooksDir string) {
	t.Helper()
	initGitRepoOnMain(t, repoDir)
	mkdirAll(t, hooksDir)
	hook := "#!/bin/sh\n" +
		"command -v git-lfs >/dev/null 2>&1 || { " +
		`printf >&2 "\n%s\n\n" "This repository is configured for Git LFS but 'git-lfs' was not found on your path."; ` +
		"exit 2; }\n" +
		"exit 0\n"
	writeFile(t, filepath.Join(hooksDir, "post-checkout"), hook)
	if err := os.Chmod(filepath.Join(hooksDir, "post-checkout"), 0755); err != nil {
		t.Fatalf("chmod post-checkout: %v", err)
	}
	runGitIsolated(t, repoDir, "config", "core.hookspath", hooksDir)
	runGitIsolated(t, repoDir, "config", "filter.lfs.required", "true")
}

func initNeutralCwd(t *testing.T, workRoot, name string) string {
	t.Helper()
	dir := filepath.Join(workRoot, name)
	mkdirAll(t, dir)
	return dir
}
```