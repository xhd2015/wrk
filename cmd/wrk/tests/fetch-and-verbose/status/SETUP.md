# Scenario

**Feature**: wrk --status Remote: field behavior on main vs linked checkout cwd

```
main repo cwd -> root Dir: . block gains Remote:
linked wt cwd -> no Remote: on any block
```

## Steps

- Descendants under `status/remote/` exercise `Remote:` presence rules.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```