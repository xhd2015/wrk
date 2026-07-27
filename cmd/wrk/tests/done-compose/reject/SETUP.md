# Scenario

**Feature**: illegal flag combinations with primary are rejected

```
wrk --done|--merge-back --json -> non-zero; --json not valid with primary

# P2: gen-commit with primary requires --commit; composed --dir rejected
wrk --gen-commit-msg --done
wrk --gen-commit-msg --model=m --done
  -> non-zero; requires --commit with primary; no primary run
wrk --gen-commit-msg --commit --dir DIR --done
  -> non-zero; --dir not valid when primary sets workDir
```

## Preconditions

- `--json` is not a valid post-modifier of `--done` / `--merge-back`.
- Bare `wrk --push` is a **standalone mode** (see `cmd/wrk/tests/push/`); not rejected here.
- **P2**: when `--gen-commit-msg` is composed with `--done`/`--merge-back`, `--commit` is **required**.
  Missing → clear error, non-zero, no primary. Prefer wording that names `--commit` and the primary
  (not only generic mutually exclusive). Composed `--dir` with primary is rejected (wrk workDir wins).

## Steps

- Leaves set the illegal combo and assert non-zero + stderr policy.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: illegal combos still run from a git main repo when useful.
	skipIfNoGit(t)
	return nil
}
```
