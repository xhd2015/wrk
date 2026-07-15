# Scenario

**Feature**: one-arg task-like + `-y` auto-promotes create from cwd

```
(cd myrepo && wrk "fix the login bug" -y)
  -> promote to --task; WRK_HOME create
```

## Steps

- Append `-y` to Args.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "-y")
	return nil
}
```
