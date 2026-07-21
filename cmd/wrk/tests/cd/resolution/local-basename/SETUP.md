# Scenario

**Feature**: wrk --cd myrepo prefers existing local directory over projects lookup

```
workspace/myrepo/ exists (non-git dir)
saved/myrepo also recorded in projects.json
channel open; cwd=workspace
wrk --cd myrepo -> follow-up uses local workspace/myrepo abs (not saved)
```

## Steps

1. Record a different saved `myrepo` under `saved/`.
2. Create local `./myrepo` under workspace cwd.
3. In-place `wrk --cd myrepo` so we can assert follow-up without shell.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	enableInPlaceChannel(t, req)
	saved := initSavedGitRepo(t, req.WorkRoot, "saved", cdBasename)
	recordSavedProject(t, req, saved)
	req.SecondRepo = resolvePath(t, saved)

	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	local := filepath.Join(req.RepoDir, cdBasename)
	mkdirAll(t, local)
	req.MainRepo = resolvePath(t, local)
	setCDFlagThenPath(req, cdBasename)
	return nil
}
```
