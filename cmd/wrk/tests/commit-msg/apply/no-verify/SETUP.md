# Scenario

**Feature**: --commit -m with --no-verify skips failing pre-commit hook

```
repo/ (hook exits 1) + staged
  -> wrk --commit -m "feat: skip hooks" --no-verify
  -> exit 0
  -> HEAD subject = "feat: skip hooks"
```

## Preconditions

- Git repo with a pre-commit hook that always exits 1.
- One staged text file.

## Steps

1. Stage change in repo with failing pre-commit.
2. Run `wrk --commit -m "feat: skip hooks" --no-verify`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	stageOneTextFileWithFailingPreCommit(t, req)
	req.Args = []string{"--commit", "-m", "feat: skip hooks", "--no-verify"}
	return nil
}
```
