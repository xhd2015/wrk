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
- `UseMinimalPath` runs wrk with `PATH=/usr/bin:/bin` and `HOME={FakeHome}`.

## Steps

- Configure the source repo with `filter.lfs.required=true` and a `core.hookspath` post-checkout hook that errors when `git-lfs` is missing (mirrors real global hooks).
- Run `wrk` with stripped PATH; the hook fails because wrk does not augment PATH for git subprocesses.

```go
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
	return home
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