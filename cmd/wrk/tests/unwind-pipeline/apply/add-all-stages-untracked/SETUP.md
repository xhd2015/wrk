# Scenario

**Feature**: apply gen-commit with `--add-all` stages untracked into the commit

```
# dirty linked dep with untracked change.txt (seedLinkedDep)
linked dep + --add-all + gen-commit + merge-back…
  -> git add -A then generate/commit
  -> untracked ends up in the landed commit on dep main
```

## Steps

1. Seed linked dirty dependency (`change.txt` untracked).
2. Run apply unwind with gen-commit + **`--add-all`** + merge-back + release flags.
3. Assert dep main HEAD contains `change.txt`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedLinkedDep(t, req)
	req.UntrackedName = "change.txt"
	req.Args = unwindGenCommitArgs(t, req, "--add-all", "--merge-back", "--sync", "--tag-next", "--push")
	return nil
}
```