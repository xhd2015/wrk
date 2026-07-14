# Scenario

**Feature**: printed script defines wrk() wrapper and follow-up env

```
wrk --bash-integration
  -> stdout contains wrk() function
  -> mentions WRK_FOLLOWUP_FILE and WRK_AUTO_CD
  -> still registers complete -o default -F _wrk wrk
```

## Steps

1. Run print-script with default isolated env.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "print")
	return nil
}
```
