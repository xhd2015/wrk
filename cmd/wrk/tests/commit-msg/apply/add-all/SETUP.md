# Scenario

**Feature**: --add-all with untracked file stages then commits with -m

```
repo/ (untracked change.go) -> wrk --commit -m "feat: add all" --add-all
  -> exit 0
  -> HEAD subject = "feat: add all"
  -> change.go is in the commit (not left untracked)
```

## Preconditions

- Isolated git repo with untracked `change.go` (not staged).

## Steps

1. Place untracked text file via `placeUntrackedTextFile`.
2. Run `wrk --commit -m "feat: add all" --add-all`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	placeUntrackedTextFile(t, req)
	req.Args = []string{"--commit", "-m", "feat: add all", "--add-all"}
	return nil
}
```
