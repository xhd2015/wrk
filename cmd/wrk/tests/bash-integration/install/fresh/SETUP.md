# Scenario

**Feature**: fresh install writes bash.sh and dual profile markers

```
empty fake HOME
wrk --bash-integration --install -> bash.sh + one marker in each profile
```

## Steps

1. Run install on empty fake HOME and isolated WRK_HOME.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "install")
	requireNoPreseed(t, req)
	return nil
}
```