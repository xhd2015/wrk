# Scenario

**Feature**: ambiguous --dep basename in non-TTY errors with candidate list

```
aaa/mydep + zzz/mydep saved
non-TTY wrk --dep mydep -> error listing both absolute paths; no external worktree
```

## Steps

1. Create consumer git repo requiring `example.com/dep`.
2. Create and record `{WorkRoot}/aaa/mydep` and `{WorkRoot}/zzz/mydep`.
3. Run `wrk --dep mydep` without `WRK_BASENAME_CONFIRM`.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerForDepBasename(t, req.WorkRoot)
	depA := initSavedDepRepo(t, req.WorkRoot, "aaa", "mydep")
	depZ := initSavedDepRepo(t, req.WorkRoot, "zzz", "mydep")
	recordSavedProject(t, req, depA)
	recordSavedProject(t, req, depZ)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = depA
	req.SecondRepo = depZ
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", "mydep"}
	return nil
}
```