# Scenario

**Feature**: fallback wrk --cd /existing/abs prints path, warns, launches shell

```
WRK_FOLLOWUP_FILE unset
fake bash on PATH
wrk --cd /WorkRoot/jumpto
  -> stdout /WorkRoot/jumpto\n
  -> stderr mentions wrk --bash-integration --install
  -> fake shell cwd = jumpto; exit 0
```

## Steps

1. Create abs target; install fake bash (exit 0).
2. Run `wrk --cd <abs>` with channel closed.

```go
func Setup(t *testing.T, req *Request) error {
	target := cdAbsTarget(t, req, "jumpto")
	req.MainRepo = target
	installFakeBash(t, req, 0)
	setCDFlagThenPath(req, target)
	return nil
}
```
