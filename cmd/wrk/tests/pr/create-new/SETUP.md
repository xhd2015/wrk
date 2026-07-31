# Scenario

**Feature**: bare `--pr` creates a new GitHub PR when none is open for the head branch

```
# linked wt + no open PR
linked wt + github origin + fake gh (list empty)
  -> wrk --pr --title T --comment C
  -> ensure remote head if missing
  -> gh pr create (body = comment)
  -> stdout: [pushed …] PR created / title set / body set / URL
```

## Steps

- Leaves choose remote-present vs remote-missing, install fake gh, set default `--pr` argv.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	_ = req
	return nil
}
```
