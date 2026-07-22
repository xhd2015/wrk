# Scenario

**Feature**: nested linked worktree under consumer still cascades via
`scan_repo.Scan` (`ListWorktrees: true`; targets from top-level worktree rows
and/or inner `Repo.Worktrees`)

```
# layout
consumer linked wt/
  deps/foo/          # git -C foodep worktree add (linked wt of foodep main)
                     # under consumerTop — cascade target

# under test
consumer wt + deps/foo linked (clean, already-included on dep)
  -> wrk --done
  -> Scan discovers deps/foo (ListWorktrees + inner Worktrees / FS walk)
  -> cascade removes deps/foo from foodep worktree list
  -> own merge-back removes consumer wt
  -> exit 0; stdout "merged branch"
```

## Preconditions

- Same behavioral fixture as `done/cascade-non-external-linked/`; this leaf lives
  under `cascade-scan-delegation/` to pin the P2 discovery contract (Scan with
  worktrees + inner field) without regressing non-external cascade.
- May already be **GREEN** if current code only walks top-level `Repos` and FS
  walk emits `deps/foo` as `RepoTypeWorktree` — still required as wiring guard
  when implementer switches collection to include `repo.Worktrees`.
- Git + Go available.

## Steps

1. Group helper: consumer wt + `deps/foo` linked worktree + clean ignore commit.
2. Run `wrk --done` with confirm-from-stdin (own may need empty confirm; cascade
   already-included removes without merge).

```go
func Setup(t *testing.T, req *Request) error {
	setupCascadeScanInnerWorktree(t, req)
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```
