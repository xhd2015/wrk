# Scenario

**Feature**: nested **main** repo under consumer → hard `Error:`, abort `--done`, no mutations (D1)

```
# consumer wt + clean external + nested full main under vendor/
  -> wrk --done
  -> non-zero
  -> stderr: Error: … nested main … <path>
  -> external + consumer still present (no cascade/own mutations)
  -> must NOT only warn+skip nested main and continue
```

## Steps

1. Build clean contained external + nested main via `setupCascadePreflightWithNestedMain`.
2. Run bare `wrk --done` from consumer worktree.

```go
func Setup(t *testing.T, req *Request) error {
	setupCascadePreflightWithNestedMain(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
