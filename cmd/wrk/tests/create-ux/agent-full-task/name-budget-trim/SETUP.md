# Scenario

**Feature**: long basename + long `-t` + `--open-in-agent` → agent gets full task text

```
basename 180×'r'; task = long prose; --open-in-agent
  -> path Base ≤255 (slug fitted)
  -> agent-run prompt = /brainstorm <full original taskDesc>
```

## Steps

1. Init long-basename repo on main.
2. Install create-ux mocks (darwin hermetic).
3. Run with full TaskDesc + `--open-in-agent` (TaskFlag `-t`).

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	basename := strings.Repeat("r", 180)
	mainRepo := filepath.Join(req.WorkRoot, basename)
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/longrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	req.RepoDir = mainRepo

	req.TaskDesc = "explore the integration of distributed tracing with opentelemetry across all microservices and platforms"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent"}
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
