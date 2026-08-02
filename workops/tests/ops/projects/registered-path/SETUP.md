# Scenario

**Feature**: ListProjects includes a path registered in temp wrk home

```
# seed main; write projects.json under WrkHome
Caller -> ListProjects(WrkHome) -> includes MainRepo path
```

## Steps

1. Seed a main repository.
2. Write `{WrkHome}/projects.json` with that path.
3. Run ListProjects with explicit WrkHome (no default HOME).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMainRepo(t, req, "myrepo")
	writeProjectsJSON(t, req.WrkHome, []string{req.MainRepo})
	return nil
}
```
