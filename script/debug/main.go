// Command debug hosts local harnesses for manual wrk verification.
//
//	go run ./script/debug create-unwind [--root DIR] [--force]
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultRoot = "/tmp/wrk-create-unwind"

	modSPL       = "example.com/spl"
	modShared    = "example.com/spl/pkgs/shared"
	modAgentPro  = "example.com/agent-pro"
	modAgentCmd  = "example.com/agent-pro/cmd"
	tagBaseline  = "v0.0.1"
	tagNext      = "v0.0.2"
	tagShared    = "pkgs/shared/v0.0.1"
	tagAgentCmd  = "cmd/v0.0.1"
	taskSlug     = "create-unwind"
	gitUserName  = "wrk-debug"
	gitUserEmail = "wrk-debug@example.com"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printRootHelp(os.Stdout)
		return fmt.Errorf("missing command")
	}
	switch args[0] {
	case "-h", "--help", "help":
		printRootHelp(os.Stdout)
		return nil
	case "create-unwind":
		return runCreateUnwind(args[1:])
	default:
		printRootHelp(os.Stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printRootHelp(w io.Writer) {
	fmt.Fprint(w, `Usage: go run ./script/debug <command> [flags]

Commands:
  create-unwind   build a local two-repo unwind fixture under /tmp

Run go run ./script/debug <command> -h for command help.
`)
}

func printCreateUnwindHelp(w io.Writer) {
	fmt.Fprint(w, `Usage: go run ./script/debug create-unwind [flags]

Create a local two-repo unwind fixture under /tmp (bare origins, wrk worktree,
external/agent-pro worktree, dirty WIP on both).

Mirrors the spl ← agent-pro crime-scene shape with nested modules
(pkgs/shared, cmd/) and isolated WRK_HOME. Origins are local bare dirs so
git push / --push never touches a remote host.

Flags:
  --root DIR     fixture root (default: /tmp/wrk-create-unwind)
  --force        remove existing root and recreate
  -h, --help     show help
`)
}

type createUnwindFlags struct {
	root  string
	force bool
}

func parseCreateUnwindFlags(args []string) (createUnwindFlags, error) {
	f := createUnwindFlags{root: defaultRoot}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			printCreateUnwindHelp(os.Stdout)
			os.Exit(0)
		case a == "--force":
			f.force = true
		case a == "--root":
			if i+1 >= len(args) {
				return f, fmt.Errorf("--root requires a directory argument")
			}
			i++
			f.root = args[i]
		case strings.HasPrefix(a, "--root="):
			f.root = strings.TrimPrefix(a, "--root=")
		default:
			return f, fmt.Errorf("unknown flag %q", a)
		}
	}
	if strings.TrimSpace(f.root) == "" {
		return f, fmt.Errorf("--root must not be empty")
	}
	return f, nil
}

func runCreateUnwind(args []string) error {
	flags, err := parseCreateUnwindFlags(args)
	if err != nil {
		return err
	}
	root, err := filepath.Abs(flags.root)
	if err != nil {
		return fmt.Errorf("resolve --root: %w", err)
	}

	if st, err := os.Stat(root); err == nil {
		if !st.IsDir() {
			return fmt.Errorf("%s exists and is not a directory", root)
		}
		if !flags.force {
			return fmt.Errorf("%s exists (pass --force to recreate)", root)
		}
		// GOMODCACHE files are often mode 0444; unlock before wipe.
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			_ = os.Chmod(path, info.Mode()|0o200)
			return nil
		})
		if err := os.RemoveAll(root); err != nil {
			return fmt.Errorf("remove %s: %w", root, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	wrkBin, err := exec.LookPath("wrk")
	if err != nil {
		return fmt.Errorf("wrk not found on PATH (install with: go install ./cmd/wrk)")
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found on PATH")
	}
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found on PATH")
	}

	date := time.Now().Format("2006-01-02")
	origins := filepath.Join(root, "origins")
	mains := filepath.Join(root, "mains")
	wrkHome := filepath.Join(root, "wrk-home")
	modproxy := filepath.Join(root, "modproxy")
	gomodcache := filepath.Join(root, "gomodcache")

	for _, d := range []string{origins, mains, wrkHome, modproxy, gomodcache} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	fmt.Println("[1/6] origins + mains")
	splBare := filepath.Join(origins, "spl.git")
	agentBare := filepath.Join(origins, "agent-pro.git")
	if err := gitBareInit(splBare); err != nil {
		return err
	}
	if err := gitBareInit(agentBare); err != nil {
		return err
	}
	fmt.Printf("      %s\n", splBare)
	fmt.Printf("      %s\n", agentBare)

	agentMain := filepath.Join(mains, "agent-pro")
	splMain := filepath.Join(mains, "spl")

	fmt.Println("[2/6] initial commit + tag v0.0.1 + push")
	if err := seedAgentPro(agentMain, agentBare, modproxy, gomodcache); err != nil {
		return fmt.Errorf("seed agent-pro: %w", err)
	}
	if err := seedSPL(splMain, splBare, modproxy, gomodcache); err != nil {
		return fmt.Errorf("seed spl: %w", err)
	}

	fmt.Println("[3/6] WRK_HOME + projects")
	if err := writeProjectsJSON(wrkHome, splMain, agentMain); err != nil {
		return err
	}
	fmt.Printf("      WRK_HOME=%s\n", wrkHome)

	fmt.Println("[4/6] wrk --new")
	wtPath, err := spawnWrkWorktree(wrkBin, wrkHome, date, splMain)
	if err != nil {
		return err
	}
	fmt.Printf("      %s\n", wtPath)

	fmt.Println("[5/6] git worktree → external/agent-pro")
	extAgent := filepath.Join(wtPath, "external", "agent-pro")
	if err := os.MkdirAll(filepath.Join(wtPath, "external"), 0o755); err != nil {
		return err
	}
	agentBranch := "main-" + date + "-" + taskSlug
	if err := runGit(agentMain, "worktree", "add", "-b", agentBranch, extAgent); err != nil {
		return fmt.Errorf("git worktree add external/agent-pro: %w", err)
	}

	fmt.Println("[6/6] dirty WIP on consumer + free")
	if err := wireReplaceAndDirty(wtPath, extAgent, modproxy, gomodcache); err != nil {
		return err
	}

	envPath := filepath.Join(root, "env.sh")
	if err := writeEnvSh(envPath, wrkHome, date, modproxy, gomodcache, wtPath); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Fixture ready.")
	fmt.Printf("  cd %s\n", wtPath)
	fmt.Printf("  source %s\n", envPath)
	fmt.Println("  wrk --unwind --show-graph")
	fmt.Println("  wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next --push --sync --reinstall-local")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Land/commit in the action graph requires --gen-commit-msg (needs an agent-runner on PATH).")
	fmt.Println("  - Bare --commit -m without --gen-commit-msg does not expand add-all/commit nodes today.")
	fmt.Println("  - origin remotes are local bare dirs under origins/; --push is safe to repeat.")
	fmt.Println("  - env.sh sets WRK_HOME, WRK_DATE, GOPROXY=file://… (includes agent-pro@v0.0.2), GOSUMDB=off, GOMODCACHE.")
	return nil
}

func seedAgentPro(dir, bare, modproxy, gomodcache string) error {
	if err := gitInitMain(dir); err != nil {
		return err
	}
	// Root module (free dep).
	if err := writeFile(filepath.Join(dir, "go.mod"),
		"module "+modAgentPro+"\n\ngo 1.22\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "agent.go"),
		"package agentpro\n\nfunc Version() string { return \""+tagBaseline+"\" }\n"); err != nil {
		return err
	}
	// Nested cmd module (agent-pro / go-pkgs cmd shape).
	cmdDir := filepath.Join(dir, "cmd")
	if err := writeFile(filepath.Join(cmdDir, "go.mod"),
		"module "+modAgentCmd+"\n\ngo 1.22\n\n"+
			"require "+modAgentPro+" "+tagBaseline+"\n\n"+
			"replace "+modAgentPro+" => ../\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(cmdDir, "main.go"),
		"package main\n\nimport (\n\t\"fmt\"\n\n\tagentpro \""+modAgentPro+"\"\n)\n\n"+
			"func main() {\n\tfmt.Println(agentpro.Version())\n}\n"); err != nil {
		return err
	}
	// Trivial install candidate for --reinstall-local.
	if err := writeFile(filepath.Join(dir, "script", "install", "main.go"),
		"package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"install agent-pro ok\") }\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "script", "install", "go.mod"),
		"module "+modAgentPro+"/script/install\n\ngo 1.22\n"); err != nil {
		return err
	}

	if err := runGit(dir, "add", "-A"); err != nil {
		return err
	}
	if err := runGit(dir, "commit", "-m", "add agent-pro modules"); err != nil {
		return err
	}
	if err := runGit(dir, "tag", tagBaseline); err != nil {
		return err
	}
	if err := runGit(dir, "tag", tagAgentCmd); err != nil {
		return err
	}
	if err := attachOriginAndPush(dir, bare); err != nil {
		return err
	}
	if err := runGit(dir, "push", "origin", tagBaseline, tagAgentCmd); err != nil {
		return err
	}
	if err := seedFileModuleProxy(modproxy, modAgentPro, tagBaseline, dir); err != nil {
		return err
	}
	if err := seedFileModuleProxy(modproxy, modAgentCmd, tagBaseline, cmdDir); err != nil {
		return err
	}
	_ = gomodcache
	return nil
}

func seedSPL(dir, bare, modproxy, gomodcache string) error {
	if err := gitInitMain(dir); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, ".gitignore"), "/external\n"); err != nil {
		return err
	}
	sharedDir := filepath.Join(dir, "pkgs", "shared")
	if err := writeFile(filepath.Join(sharedDir, "go.mod"),
		"module "+modShared+"\n\ngo 1.22\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Hello() string { return \"shared\" }\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "go.mod"),
		"module "+modSPL+"\n\ngo 1.22\n\n"+
			"require (\n"+
			"\t"+modAgentPro+" "+tagBaseline+"\n"+
			"\t"+modShared+" "+tagBaseline+"\n"+
			")\n\n"+
			"replace "+modShared+" => ./pkgs/shared\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "main.go"),
		"package main\n\nimport (\n\t\"fmt\"\n\n\tagentpro \""+modAgentPro+"\"\n\tshared \""+modShared+"\"\n)\n\n"+
			"func main() {\n\tfmt.Println(shared.Hello(), agentpro.Version())\n}\n"); err != nil {
		return err
	}

	if err := runGo(dir, modproxy, gomodcache, "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy (spl): %w", err)
	}
	if err := runGit(dir, "add", "-A"); err != nil {
		return err
	}
	if err := runGit(dir, "commit", "-m", "add spl modules"); err != nil {
		return err
	}
	if err := runGit(dir, "tag", tagBaseline); err != nil {
		return err
	}
	if err := runGit(dir, "tag", tagShared); err != nil {
		return err
	}
	if err := attachOriginAndPush(dir, bare); err != nil {
		return err
	}
	if err := runGit(dir, "push", "origin", tagBaseline, tagShared); err != nil {
		return err
	}
	if err := seedFileModuleProxy(modproxy, modSPL, tagBaseline, dir); err != nil {
		return err
	}
	if err := seedFileModuleProxy(modproxy, modShared, tagBaseline, sharedDir); err != nil {
		return err
	}
	return nil
}

func writeProjectsJSON(wrkHome string, paths ...string) error {
	type project struct {
		Path    string `json:"path"`
		AddedAt string `json:"added_at"`
		Source  string `json:"source"`
	}
	type file struct {
		Version  int       `json:"version"`
		Projects []project `json:"projects"`
	}
	now := time.Now().UTC().Format(time.RFC3339)
	pf := file{Version: 1}
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		pf.Projects = append(pf.Projects, project{Path: abs, AddedAt: now, Source: "manual"})
	}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(wrkHome, "projects.json"), data, 0o644)
}

func spawnWrkWorktree(wrkBin, wrkHome, date, splMain string) (string, error) {
	cmd := exec.Command(wrkBin, "--new", "--here", "--no-cd", "-t", taskSlug, splMain)
	cmd.Env = append(os.Environ(),
		"WRK_HOME="+wrkHome,
		"WRK_DATE="+date,
	)
	cmd.Dir = splMain
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("wrk --new: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	path := strings.TrimSpace(string(out))
	// Prefer last non-empty line (stdout should be the path alone with --here --no-cd).
	lines := strings.Split(path, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, wrkHome) || strings.Contains(line, string(filepath.Separator)+"worktrees"+string(filepath.Separator)) {
			if st, err := os.Stat(line); err == nil && st.IsDir() {
				return line, nil
			}
		}
		// Fallback: any existing directory line.
		if st, err := os.Stat(line); err == nil && st.IsDir() {
			return line, nil
		}
	}
	return "", fmt.Errorf("wrk --new did not print a worktree path; output:\n%s", strings.TrimSpace(string(out)))
}

func wireReplaceAndDirty(wtPath, extAgent, modproxy, gomodcache string) error {
	gomod := filepath.Join(wtPath, "go.mod")
	content, err := os.ReadFile(gomod)
	if err != nil {
		return err
	}
	s := string(content)
	replaceLine := "replace " + modAgentPro + " => ./external/agent-pro\n"
	if !strings.Contains(s, "=> ./external/agent-pro") {
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		s += replaceLine
		if err := os.WriteFile(gomod, []byte(s), 0o644); err != nil {
			return err
		}
	}
	if err := writeFile(filepath.Join(wtPath, "FEATURE_WIP.txt"), "consumer feature WIP\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(extAgent, "agent.go"),
		"package agentpro\n\nfunc Version() string { return \"wip-next\" }\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(extAgent, "FREE_WIP.txt"), "free dep feature WIP\n"); err != nil {
		return err
	}
	// Keep go.sum coherent with the replace overlay.
	if err := runGo(wtPath, modproxy, gomodcache, "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy (worktree): %w", err)
	}
	// Pre-seed the likely next free tag so offline pin/tidy after --tag-next
	// can resolve without a network module proxy.
	if err := seedFileModuleProxy(modproxy, modAgentPro, tagNext, extAgent); err != nil {
		return fmt.Errorf("seed agent-pro@%s modproxy: %w", tagNext, err)
	}
	return nil
}

func writeEnvSh(path, wrkHome, date, modproxy, gomodcache, wtPath string) error {
	absProxy, err := filepath.Abs(modproxy)
	if err != nil {
		return err
	}
	absCache, err := filepath.Abs(gomodcache)
	if err != nil {
		return err
	}
	body := fmt.Sprintf(`# sourced by: source %s
export WRK_HOME=%q
export WRK_DATE=%q
export GOPROXY=%q
export GOSUMDB=off
export GONOSUMDB=*
export GOMODCACHE=%q
# optional: cd to the consumer worktree
# cd %q
`, path, wrkHome, date, "file://"+absProxy, absCache, wtPath)
	return os.WriteFile(path, []byte(body), 0o644)
}

// --- git / go / filesystem helpers ---

func gitBareInit(path string) error {
	cmd := exec.Command("git", "-c", "init.templateDir=", "init", "--bare", "-b", "main", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init --bare %s: %w\n%s", path, err, out)
	}
	return nil
}

func gitInitMain(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := runGit(dir, "-c", "init.templateDir=", "init", "-b", "main"); err != nil {
		return err
	}
	if err := runGit(dir, "config", "user.email", gitUserEmail); err != nil {
		return err
	}
	if err := runGit(dir, "config", "user.name", gitUserName); err != nil {
		return err
	}
	// Isolate from global/user hooks (e.g. wip-protect) so --push is repeatable.
	if err := runGit(dir, "config", "core.hooksPath", "/dev/null"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "README.md"), "# "+filepath.Base(dir)+"\n"); err != nil {
		return err
	}
	if err := runGit(dir, "add", "README.md"); err != nil {
		return err
	}
	return runGit(dir, "commit", "-m", "init")
}

func attachOriginAndPush(repo, bare string) error {
	if err := runGit(repo, "remote", "add", "origin", bare); err != nil {
		return err
	}
	return runGit(repo, "push", "-u", "origin", "main")
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+gitUserName,
		"GIT_AUTHOR_EMAIL="+gitUserEmail,
		"GIT_COMMITTER_NAME="+gitUserName,
		"GIT_COMMITTER_EMAIL="+gitUserEmail,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s (in %s): %w\n%s", strings.Join(args, " "), dir, err, out)
	}
	return nil
}

func runGo(dir, modproxy, gomodcache string, args ...string) error {
	absProxy, err := filepath.Abs(modproxy)
	if err != nil {
		return err
	}
	absCache, err := filepath.Abs(gomodcache)
	if err != nil {
		return err
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOPROXY=file://"+absProxy,
		"GOSUMDB=off",
		"GONOSUMDB=*",
		"GOMODCACHE="+absCache,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func seedFileModuleProxy(proxyRoot, modulePath, version, srcDir string) error {
	parts := strings.Split(modulePath, "/")
	vDir := filepath.Join(append([]string{proxyRoot}, parts...)...)
	vDir = filepath.Join(vDir, "@v")
	if err := os.MkdirAll(vDir, 0o755); err != nil {
		return err
	}
	modContent, err := os.ReadFile(filepath.Join(srcDir, "go.mod"))
	if err != nil {
		return err
	}
	// Proxy .mod must not carry local replace directives.
	modClean := stripReplaceDirectives(string(modContent))
	if err := os.WriteFile(filepath.Join(vDir, version+".mod"), []byte(modClean), 0o644); err != nil {
		return err
	}
	info := fmt.Sprintf(`{"Version":%q,"Time":"2026-07-01T00:00:00Z"}`+"\n", version)
	if err := os.WriteFile(filepath.Join(vDir, version+".info"), []byte(info), 0o644); err != nil {
		return err
	}
	listPath := filepath.Join(vDir, "list")
	existing := ""
	if data, err := os.ReadFile(listPath); err == nil {
		existing = string(data)
	}
	if !strings.Contains(existing, version) {
		if err := os.WriteFile(listPath, []byte(existing+version+"\n"), 0o644); err != nil {
			return err
		}
	}
	return writeModuleZip(filepath.Join(vDir, version+".zip"), modulePath, version, srcDir)
}

func stripReplaceDirectives(mod string) string {
	var out []string
	lines := strings.Split(mod, "\n")
	inReplaceBlock := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if inReplaceBlock {
			if trim == ")" {
				inReplaceBlock = false
			}
			continue
		}
		if strings.HasPrefix(trim, "replace (") {
			inReplaceBlock = true
			continue
		}
		if strings.HasPrefix(trim, "replace ") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func writeModuleZip(zipPath, modulePath, version, srcDir string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	prefix := modulePath + "@" + version + "/"
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if path != srcDir && (base == ".git" || base == "external") {
				return filepath.SkipDir
			}
			if path != srcDir {
				if _, statErr := os.Stat(filepath.Join(path, "go.mod")); statErr == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			return nil
		}
		rel = filepath.ToSlash(rel)
		w, err := zw.Create(prefix + rel)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, in)
		_ = in.Close()
		return copyErr
	})
	if err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}
