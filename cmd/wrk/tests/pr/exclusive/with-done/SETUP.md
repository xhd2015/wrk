# Scenario

**Feature**: `--pr` is mutually exclusive with `--done`

```
myrepo -> wrk --pr --title T --comment C --done
  -> non-zero
  -> stderr indicates mutual exclusion / mode conflict
```

## Steps

1. Seed minimal main repo.
2. Run `--pr` with title/comment and `--done`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrExclusiveMinimal(t, req)
	req.Args = append(prDefaultArgs(), "--done")
	return nil
}
```
