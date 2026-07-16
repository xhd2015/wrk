# Scenario

**Feature**: `--tag-next --push --dry-run` plans tags and would-push branch + tags; zero mutations

```
# root bump + origin
git repo + bare origin
  -> wrk --tag-next --push --dry-run
  -> plan v0.0.2; 1 tag planned
  -> would: git push origin main
  -> would: git push origin v0.0.2
  -> no local tag; origin unchanged
```

## Steps

1. `setupPushRepo`.
2. Snapshot origin main SHA + tags for Assert.
3. Run `wrk --tag-next --push --dry-run`.

```go
import (
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	setupPushRepo(t, req)
	sha := strings.TrimSpace(gitOutputIsolated(t, req.OriginBare, "rev-parse", "refs/heads/main"))
	writeFile(t, filepath.Join(req.WorkRoot, "origin-main-before"), sha+"\n")
	// Origin may have no tags yet (show-ref --tags exits 1 when empty).
	// Only snapshot main tip; Assert checks remoteTagExists is false for next tag.
	req.Args = []string{"--tag-next", "--push", "--dry-run"}
	return nil
}
```