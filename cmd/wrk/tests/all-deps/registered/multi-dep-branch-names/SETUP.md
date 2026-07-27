# Scenario

**Feature**: multi-dep both on main → branch is `{token}-{date}` without dep basename (P2 E2)

```
# consumer requires dep1+dep2; both registered on main
# --all-deps links both; each dep repo owns branch main-{date} (no mydepN- prefix)
# paths stay mydep1-main-{date} and mydep2-main-{date} (separate basenames; no path collision)
# Confirmed E2: separate dep repos both get unsuffixed main-{date} (branch lives per depMain).
# Do not force artificial -1 across separate repos. Pre-existing branch → -1: dep/branch-collision-suffix.
```


## Steps

1. Create dep repos `mydep1` and `mydep2` on `main`.
2. Register both via `wrk --add`.
3. Create consumer requiring both modules.
4. Run `wrk --all-deps` from the consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	registeredEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	initAllDepsRepo(t, dep2, "example.com/dep2", "dep2")
	registerAllDepsProjects(t, req, dep1, dep2)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```
