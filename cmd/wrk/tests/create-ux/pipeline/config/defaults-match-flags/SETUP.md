# Scenario

**Feature**: full create config defaults without flags match full-pipeline flags

```
config window+terminal+agent on; wrk -t 'ship feature'
  -> space + iterm ForceNew + agent follow-up
```

## Steps

1. Write full create UX config.
2. Run with task only (no UX flags).

```go
func Setup(t *testing.T, req *Request) error {
	writeFullCreateUXConfig(t, req.WrkHome)
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = nil
	return nil
}
```
