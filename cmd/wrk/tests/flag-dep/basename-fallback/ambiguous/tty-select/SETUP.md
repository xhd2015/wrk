# Scenario

**Feature**: ambiguous --dep basename resolved via TTY prompt with stdin selection

```
aaa/mydep + zzz/mydep saved (both example.com/dep)
WRK_BASENAME_CONFIRM=1 + stdin "2" -> --dep succeeds from zzz/mydep (lex-sorted #2)
```

## Steps

1. Create consumer git repo requiring `example.com/dep`.
2. Create git dep repos `{WorkRoot}/aaa/mydep` and `{WorkRoot}/zzz/mydep`.
3. Record both with `wrk --add`.
4. Run `wrk --dep mydep` with `WRK_BASENAME_CONFIRM=1` and stdin `2\n`.

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
	req.SelectedSavedRepo = depZ
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", "mydep"}
	req.BasenameEnv = "WRK_BASENAME_CONFIRM=1"
	req.StdinInput = "2\n"
	return nil
}
```