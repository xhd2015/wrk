# Scenario

**Feature**: CLI flags drive create UX for this invocation

```
wrk --new-window | --new-terminal | --reuse-terminal | --smart-terminal | --open-in-agent
```

## Steps

- Empty config; flags only.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
