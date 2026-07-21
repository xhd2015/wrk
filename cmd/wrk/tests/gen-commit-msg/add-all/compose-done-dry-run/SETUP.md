# Scenario

**Feature**: compose peels `--add-all` so `--gen-commit-msg --add-all --commit --done --dry-run` plans stage-all

```
# Classic RED until genCommitMsgBoolFlags includes --add-all
linked wt (ahead) + staged-for-commit.go
  -> wrk --gen-commit-msg --add-all --commit --done --dry-run
  -> must NOT: unrecognized flag: --add-all (lessflags on remaining args)
  -> stderr: would: git add -A
  -> gen-commit dry plan (mock B / would: git commit) + primary dry plan
  -> exit 0; zero permanent mutation preferred
```

## Preconditions

- Isolated main repo + linked worktree (`feature-work` ahead of main).
- One staged text file on the worktree (`staged-for-commit.go`).
- No untracked dirt (staged-only keeps MergeBack dry-run path clean after stash).
- Classic RED while peel leaves `--add-all` for lessflags.

## Steps

1. Build linked worktree fixture with staged file (`setupLinkedWtWithStagedForCompose`).
2. Run from worktree cwd:
   `wrk --gen-commit-msg --add-all --commit --done --dry-run`

```go
func Setup(t *testing.T, req *Request) error {
	setupLinkedWtWithStagedForCompose(t, req)
	req.Args = []string{
		"--gen-commit-msg",
		"--add-all",
		"--commit",
		"--done",
		"--dry-run",
	}
	return nil
}
```
