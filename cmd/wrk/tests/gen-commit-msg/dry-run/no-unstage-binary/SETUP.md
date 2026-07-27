# Scenario

**Feature**: wrk --gen-commit-msg --dry-run plans binary unstage without mutating index

```
# stage binary + text through wrk binary; dry-run does not unstage
staged: app.go + blob.bin -> wrk --gen-commit-msg --dry-run
  -> stderr: would: unstage … blob.bin
  -> index still has blob.bin staged
  -> mock N = 2 (count before unstage)
```

## Preconditions

- Isolated git repo with hooks disabled.
- Stages one text file (`app.go`) and one ELF-like binary (`blob.bin`).

## Steps

1. Init repo; stage `app.go` and `blob.bin`.
2. Run `wrk --gen-commit-msg --dry-run` from the repo cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageBinaryAndTextFile(t, req)
	req.Args = []string{"--gen-commit-msg", "--dry-run"}
	return nil
}
```
