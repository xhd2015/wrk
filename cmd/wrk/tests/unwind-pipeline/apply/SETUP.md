# Scenario

**Feature**: applying unwind executes repository-local stages in peel order

```
dirty dependency worktree -> unwind apply
  -> gen-commit (stage only if --add-all) -> land -> post stages
```

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
