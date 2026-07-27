# Scenario

**Feature**: bare `wrk --scan-git-repos` uses universe home and product home cache files

```
HOME=FakeHome
  FakeHome/home-main
  FakeHome/Projects/proj-main
  -> wrk --scan-git-repos   # default root ~
  -> CacheRoot = $HOME/.cache/git-repo-scan
  -> universe home: home/repos.json present
  -> projects.json empty; stdout includes both
```

## Preconditions

- Parent FakeHome fixtures: `home-main` + `Projects/proj-main`.
- No explicit ROOT args (default root is `$HOME`).

## Steps

1. Set Args to bare `--scan-git-repos`.
2. Force `WRK_SCAN_DEBUG=` so ambient host debug does not pollute.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--scan-git-repos"}
	req.ExtraEnv = append(req.ExtraEnv, "WRK_SCAN_DEBUG=")
	return nil
}
```
