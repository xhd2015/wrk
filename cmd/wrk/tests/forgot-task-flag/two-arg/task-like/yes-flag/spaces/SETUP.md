# Scenario

**Feature**: two-arg spaces + `-y` → task create under WRK_HOME

```
wrk <myrepo> "fix the login bug" -y
  -> promoted create; slug in path/branch
```

## Steps

1. Two-arg multi-word second positional.
2. Parent group already appends `-y`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupTwoArg(t, req, taskLikeSpaces)
	return nil
}
```
