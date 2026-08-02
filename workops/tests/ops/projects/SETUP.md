# Scenario

**Feature**: ListProjects reads registered mains from an injectable wrk home

```
# temp wrkHome + projects.json
Caller -> workops.ListProjects(wrkHome) -> []Project
```

## Steps

1. Grouping only: set Op to list-projects.
2. Leaves write projects.json under req.WrkHome.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpListProjects
	return nil
}
```
