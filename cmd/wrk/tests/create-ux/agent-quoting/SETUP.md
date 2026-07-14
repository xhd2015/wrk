# Scenario

**Feature**: adversarial task characters stay argv/shell-safe for agent

```
wrk -t 'fix "quoted" task' --open-in-agent
  -> prompt correctly represented; no broken shell metacharacters
```

## Steps

- Leaves use quoted task text with agent paths.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
