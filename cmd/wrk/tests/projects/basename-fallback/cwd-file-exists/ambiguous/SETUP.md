# Scenario

**Feature**: file in cwd plus multiple saved projects yields ambiguous guided hint

```
workspace/spl (file) + aaa/spl + zzz/spl in projects.json
wrk spl <flags> -> guided stderr listing all matches + <full-path> hint
```

## Steps

- Descendants seed two saved repos sharing basename with the cwd file name.
- Run `wrk <basename>` with mode flags; hint preserves flags and uses `<full-path>` placeholder.

```go
func Setup(t *testing.T, req *Request) error {
	ensureCwdFileExistsHelpersUsed()
	return nil
}
```