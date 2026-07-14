# Scenario

**Feature**: <target-dir> exists on disk — split on dir vs file and collision behavior

```
# <target-dir> exists as a dir -> spawn default-named sub-dir under it (with -N on collision)
wrk <dir> <existing-dir> -> <existing-dir>/<basename>-<token>-<date>[-N]
# <target-dir> exists as a file -> error (not a directory)
```

## Preconditions

- Leaves pre-create `<target-dir>` under `{WorkRoot}` before running `wrk`.

## Steps

- `basic-subdir/` pre-creates an empty `{WorkRoot}/target` dir.
- `collision-suffix/` pre-creates `{WorkRoot}/target` plus the would-be sub-dir
  `{WorkRoot}/target/myrepo-main-2026-06-30` to force the `-1` suffix.
- `target-is-file/` pre-creates `{WorkRoot}/target` as a regular file.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
