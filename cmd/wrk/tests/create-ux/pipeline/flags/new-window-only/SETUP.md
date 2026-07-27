# Scenario

**Feature**: `--new-window` alone implies terminal mode new; no agent

```
wrk --new-window
  -> create; space CreateAndActivate once; iterm ForceNew at wt path; no agent
```

## Steps

1. Run `wrk --new-window`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--new-window"}
	return nil
}
```
