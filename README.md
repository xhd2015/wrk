# wrk

Git worktree helper for isolated feature branches. Create linked worktrees, manage
dependencies, merge back, and inspect project status — without disturbing your
main checkout.

## Install

```sh
go install github.com/xhd2015/wrk/cmd/wrk@latest
```

Requires [Go](https://go.dev/) 1.25+ and `git` on `PATH`.

## Quick start

```sh
wrk                              # create worktree from cwd
wrk -t 'fix login bug'           # append task slug to branch/dir names
wrk --done                       # merge back and remove worktree
```

## Common commands

| Command | Purpose |
|---------|---------|
| `wrk --status` | Status for git repos under this checkout |
| `wrk -l` | List worktrees |
| `wrk --projects` | Recorded main repository paths |
| `wrk --projects --github` | Same as `--projects`, only github.com origin remotes |
| `wrk --where <basename>` | Look up saved project path(s) |
| `wrk --main` | Nested shell at main repository root |
| `wrk --dep <path>` | Spawn a dependency worktree under `./external` |
| `wrk --bring <path>` | Like `--dep`; soft-skip replace when not a module dep |
| `wrk --all-deps` | Link required deps from registered projects |
| `wrk --web` | Local web UI (React SPA + API on 127.0.0.1) |

Run `wrk -h` for the full flag list.

## Configuration

- **`WRK_HOME`** — storage root (default: `~/.wrk`). Holds worktrees, `projects.json`,
  `events.jsonl`, and `config.json`.
- **`WRK_DATE`** — override the run date (`YYYY-MM-DD`) used in worktree/branch names.
- **`wrk --set-config`** — manage create UX defaults (iTerm2, Mission Control, agent launch).

Bash tab-completion and auto-cd: `wrk --bash-integration`.

## Dependencies

`wrk` is a standalone CLI that reuses shared libraries from
[dot-pkgs](https://github.com/xhd2015/dot-pkgs):

- **Go** — `github.com/xhd2015/dot-pkgs/go-pkgs` (git, gotool, shell, pathfmt, …)
- **React** — `dot-pkgs/react` shared components (`routePrefix`, API client) consumed by
  `wrk-react/`

Local development expects sibling checkouts:

```
$X/
├── wrk/
└── dot-pkgs/
```

`go.mod` includes `replace github.com/xhd2015/dot-pkgs/go-pkgs => ../dot-pkgs/go-pkgs`.

## Development

```sh
# Build CLI
go build -o wrk ./cmd/wrk

# Unit tests
go test ./...

# Doctest integration suite
doctest test -v ./...

# Rebuild embedded web UI
./script/build-frontend.sh

# Dev server with HMR (requires bun)
wrk --web --dev
```

### Project layout

```
cmd/wrk/          CLI entry + doctest tree
wrkcli/           Core logic, storage, wrkserver, embedded web/dist
wrk-react/        wrk web SPA (depends on dot-pkgs/react)
script/           Frontend build helper
```

## License

MIT — see [LICENSE](LICENSE).