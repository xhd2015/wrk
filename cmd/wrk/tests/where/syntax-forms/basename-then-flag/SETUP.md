# Scenario

**Feature**: wrk BASE --where form (basename then flag)

```
saved/spl recorded
workspace/ -> wrk spl --where -> stdout saved abs path
```

## Steps

1. Parent recorded `spl` and set neutral cwd.
2. TargetDir=spl, Args=`["--where"]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setWhereBasenameThenFlag(req, whereBasename)
	return nil
}
```
