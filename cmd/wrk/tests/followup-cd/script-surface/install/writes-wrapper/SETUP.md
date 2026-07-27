# Scenario

**Feature**: fresh install writes bash.sh containing wrk() + follow-up logic

```
empty WRK_HOME
wrk --bash-integration --install
  -> bash.sh defines wrk() and WRK_FOLLOWUP_FILE handling
```

## Steps

1. Run install with no pre-seeded bash.sh.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "install")
	if req.PreExistingBashSh != "" {
		t.Fatalf("expected no pre-seeded bash.sh")
	}
	return nil
}
```
