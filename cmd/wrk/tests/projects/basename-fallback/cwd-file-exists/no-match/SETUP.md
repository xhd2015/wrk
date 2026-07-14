# Scenario

**Feature**: file in cwd with no saved project yields short error only

```
workspace/foo (file), no projects.json match
wrk foo -> single-line stderr: file exists; no registry block or hint
```

## Steps

- Descendants create cwd file `./<basename>` with no matching saved project.
- Run create-mode `wrk <basename>`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureCwdFileExistsHelpersUsed()
	return nil
}
```