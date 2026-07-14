# Scenario

**Feature**: status reports installed when script and both markers are present

```
real install into isolated env
wrk --bash-integration --status -> installed, exit 0
```

## Steps

1. Run real install before status (`req.PreInstall = true`).

```go
func Setup(t *testing.T, req *Request) error {
	req.PreInstall = true
	return nil
}
```