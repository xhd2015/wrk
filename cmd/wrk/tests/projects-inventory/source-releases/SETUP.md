# Scenario

**Feature**: ResolveSourceReleases maps numeric git tags to go require versions

```
# Op = source-releases
sourceMain (git repo with modules)
  -> wrkcli.ResolveSourceReleases(sourceMain)
  -> Releases[{ModulePath, Tag, Version}] + Missing[]
```

## Preconditions

- Leaves under this branch set `req.Op = OpSourceReleases` and `req.SourceMain`.
- Tag mapping: root `vX.Y.Z` → version `vX.Y.Z`; nested `sub/vX.Y.Z` → version `vX.Y.Z`.
- Only numeric release tags count (no prerelease).

## Steps

1. Set operation to source-releases.
2. Leaves create a multi-module source repo, apply tags, set WantReleases/WantMissing.
3. Run calls ResolveSourceReleases(SourceMain).

## Context

- Does not read WRK_HOME projects.json (source path is explicit).
- Missing tags are per-module soft omissions into Missing, not a hard inventory error.

```go
func Setup(t *testing.T, req *Request) error {
	req.Op = OpSourceReleases
	return nil
}
```
