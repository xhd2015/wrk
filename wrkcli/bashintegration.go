package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xhd2015/wrk/wrkcli/storage"
)

const (
	wrkMarkerBegin = "# === wrk integration begin ==="
	wrkMarkerEnd   = "# === wrk integration end ==="
)

const bashIntegrationScript = `#!/usr/bin/env bash
# wrk bash tab completion + auto-cd wrapper (installed at ${WRK_HOME:-$HOME/.wrk}/integration/bash.sh)

_wrk() {
  local cur
  cur="${COMP_WORDS[COMP_CWORD]}"
  # Path-like cur (/ ./ ../): yield to bash default filename completion.
  if [[ "$cur" == /* || "$cur" == ./* || "$cur" == ../* ]]; then
    COMPREPLY=()
    compopt -o default 2>/dev/null
    return
  fi
  local candidates
  candidates=$(command wrk --bash-integration --complete -- "${COMP_WORDS[@]}" "$COMP_CWORD")
  COMPREPLY=()
  if [[ -n "$candidates" ]]; then
    while IFS= read -r line; do
      [[ -n "$line" ]] && COMPREPLY+=("$line")
    done <<< "$candidates"
  fi
}

# wrk() wrapper: when auto-cd is enabled, set WRK_FOLLOWUP_FILE so the binary
# can append "cd /abs" lines; print each to stderr and builtin-cd after exit.
wrk() {
  local _wrk_skip_followup=0
  local _wrk_arg
  if [[ "${WRK_AUTO_CD:-}" == "0" ]]; then
    _wrk_skip_followup=1
  else
    for _wrk_arg in "$@"; do
      if [[ "$_wrk_arg" == "--no-cd" ]]; then
        _wrk_skip_followup=1
        break
      fi
    done
  fi

  if [[ "$_wrk_skip_followup" -eq 1 ]]; then
    command wrk "$@"
    return $?
  fi

  local _wrk_followup
  _wrk_followup="$(mktemp "${TMPDIR:-/tmp}/wrk-followup.XXXXXX")" || return 1
  export WRK_FOLLOWUP_FILE="$_wrk_followup"
  command wrk "$@"
  local _wrk_status=$?
  unset WRK_FOLLOWUP_FILE

  local _wrk_line _wrk_path _wrk_cd_failed=0
  if [[ -f "$_wrk_followup" ]]; then
    while IFS= read -r _wrk_line || [[ -n "$_wrk_line" ]]; do
      [[ -z "$_wrk_line" ]] && continue
      # Whitelist: only "cd" + a single absolute path (never eval).
      case "$_wrk_line" in
        cd\ /*)
          _wrk_path="${_wrk_line#cd }"
          # Reject paths with spaces / extra fields.
          if [[ "$_wrk_path" == *" "* || "$_wrk_path" != /* ]]; then
            continue
          fi
          printf '%s\n' "cd $_wrk_path" >&2
          if ! builtin cd "$_wrk_path"; then
            _wrk_cd_failed=1
            break
          fi
          ;;
      esac
    done < "$_wrk_followup"
  fi
  rm -f "$_wrk_followup"
  if [[ "$_wrk_cd_failed" -ne 0 ]]; then
    return 1
  fi
  return "$_wrk_status"
}

complete -o default -F _wrk wrk
`

const wrkMarkerBlock = `# === wrk integration begin ===
_wrk_home="${WRK_HOME:-$HOME/.wrk}"
[[ -f "$_wrk_home/integration/bash.sh" ]] && source "$_wrk_home/integration/bash.sh"
# === wrk integration end ===
`

var wrkCompletionFlags = []string{
	"-h", "--help",
	"-l", "--list",
	"--done",
	"--merge-back",
	"--status",
	"--repos",
	"--projects",
	"--projects-dep-graph",
	"--scan-git-repos",
	"--no-cache",
	"--include-worktrees",
	"--fetch",
	"--github",
	"-v", "--verbose",
	"--color",
	"--add",
	"--rm",
	"--where",
	"--cd",
	"--main",
	"--unwind",
	"--pin-locals",
	"--dep-replace",
	"--dep-update",
	"--bring",
	"--no-dep",
	"--tag-next",
	"--propagate-tags",
	"--pr",
	"--title",
	"--comment",
	"--sync",
	"--dry-run",
	"--commit",
	"-m", "--message",
	"--no-verify",
	"--add-all",
	"--gen-commit-msg",
	"-t", "--task",
	"--set-task",
	"-y", "--yes",
	"--confirm",
	"--confirm-from-stdin",
	"--no-in-module-replace",
	"--no-cd",
	"--force-cd",
	"--new-window",
	"--no-new-window",
	"--new-terminal",
	"--reuse-terminal",
	"--smart-terminal",
	"--no-new-terminal",
	"--open-in-agent",
	"--no-open-in-agent",
	"--no-config",
	"--set-config",
	"--create",
	"--show",
	"--bash-integration",
	"--install",
	"--uninstall",
	"--complete",
}

var bashIntegrationDisallowedFlags = []string{
	"--done", "--merge-back", "-l", "--list", "--repos", "--projects",
	"--projects-dep-graph",
	"--scan-git-repos", "--no-cache", "--include-worktrees",
	"--fetch", "--github", "-v", "--verbose", "--color", "--add", "--rm", "--where",
	"--cd",
	"--main",
	"--pin-locals",
	"--dep-replace", "--dep-update",
	"--bring", "--no-dep", "--tag-next", "--propagate-tags", "--sync", "-t", "--task", "--set-task", "-y", "--yes",
	"--confirm", "--confirm-from-stdin", "--no-in-module-replace",
}

// CompletionRequest carries bash COMP_WORDS and COMP_CWORD for tab completion.
type CompletionRequest struct {
	Words []string
	CWord int
}

// ExitCodeError signals a non-zero exit without stderr output.
type ExitCodeError struct {
	Code int
}

func (e ExitCodeError) Error() string { return "" }

func runBashIntegration(args []string) error {
	if err := checkBashIntegrationMutualExclusion(args); err != nil {
		return err
	}

	action, dryRun, completeReq, err := parseBashIntegrationArgs(args)
	if err != nil {
		return err
	}

	switch action {
	case "":
		// Use Printf so go vet does not treat the script (which contains %s) as a format string.
		fmt.Printf("%s", bashIntegrationScript)
		return nil
	case "--install":
		if dryRun {
			return installBashIntegrationDryRun()
		}
		return installBashIntegration()
	case "--uninstall":
		if dryRun {
			return uninstallBashIntegrationDryRun()
		}
		return uninstallBashIntegration()
	case "--status":
		if dryRun {
			return fmt.Errorf("wrk: unknown integration action %q", "--dry-run")
		}
		if code := statusBashIntegration(); code != 0 {
			return ExitCodeError{Code: code}
		}
		return nil
	case "--complete":
		return runBashComplete(completeReq)
	default:
		return fmt.Errorf("wrk: unknown integration action %q", action)
	}
}

func checkBashIntegrationMutualExclusion(args []string) error {
	allowed := map[string]bool{
		"--bash-integration": true,
		"--install":          true,
		"--uninstall":        true,
		"--status":           true,
		"--dry-run":          true,
		"--complete":         true,
		"-h":                 true,
		"--help":             true,
		"--":                 true,
	}

	afterCompleteSep := false
	for _, arg := range args {
		if afterCompleteSep {
			continue
		}
		if arg == "--" {
			afterCompleteSep = true
			continue
		}
		if allowed[arg] {
			continue
		}
		for _, d := range bashIntegrationDisallowedFlags {
			if arg == d {
				return fmt.Errorf("wrk: --bash-integration is mutually exclusive with other modes")
			}
		}
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("wrk: --bash-integration is mutually exclusive with other modes")
		}
		return fmt.Errorf("wrk: --bash-integration is mutually exclusive with other modes")
	}
	return nil
}

func parseBashIntegrationArgs(args []string) (action string, dryRun bool, completeReq *CompletionRequest, err error) {
	afterBashIntegration := false
	afterCompleteSep := false
	var completeWords []string

	for _, arg := range args {
		if !afterBashIntegration {
			if arg == "--bash-integration" {
				afterBashIntegration = true
			}
			continue
		}
		if afterCompleteSep {
			completeWords = append(completeWords, arg)
			continue
		}
		switch arg {
		case "--install", "--uninstall", "--status", "--complete":
			if action != "" {
				return "", false, nil, fmt.Errorf("wrk: unknown integration action %q", arg)
			}
			action = arg
		case "--dry-run":
			dryRun = true
		case "--":
			if action != "--complete" {
				return "", false, nil, fmt.Errorf("wrk: unknown integration action %q", arg)
			}
			afterCompleteSep = true
		case "-h", "--help":
			// Hidden from main help; no dedicated integration help in tests.
		default:
			return "", false, nil, fmt.Errorf("wrk: unknown integration action %q", arg)
		}
	}

	if action == "--complete" {
		if len(completeWords) < 1 {
			return "", false, nil, fmt.Errorf("wrk: --complete requires words and cword after --")
		}
		cwordStr := completeWords[len(completeWords)-1]
		cword, convErr := strconv.Atoi(cwordStr)
		if convErr != nil {
			return "", false, nil, fmt.Errorf("wrk: invalid completion cword %q", cwordStr)
		}
		words := completeWords[:len(completeWords)-1]
		completeReq = &CompletionRequest{Words: words, CWord: cword}
	}
	return action, dryRun, completeReq, nil
}

func bashIntegrationPaths() (home, wrkHome, scriptPath, bashProfilePath, bashrcPath string, err error) {
	home, err = os.UserHomeDir()
	if err != nil {
		return "", "", "", "", "", err
	}
	wrkHome, err = resolveWrkHome()
	if err != nil {
		return "", "", "", "", "", err
	}
	scriptPath = filepath.Join(wrkHome, "integration", "bash.sh")
	bashProfilePath = filepath.Join(home, ".bash_profile")
	bashrcPath = filepath.Join(home, ".bashrc")
	return home, wrkHome, scriptPath, bashProfilePath, bashrcPath, nil
}

func scriptPresent(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func markerPresent(profilePath string) bool {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), wrkMarkerBegin)
}

func fullyInstalled(home string) bool {
	return markerPresent(filepath.Join(home, ".bash_profile")) &&
		markerPresent(filepath.Join(home, ".bashrc"))
}

func fullyUninstalled(home string) bool {
	return !markerPresent(filepath.Join(home, ".bash_profile")) &&
		!markerPresent(filepath.Join(home, ".bashrc"))
}

func installBashIntegrationDryRun() error {
	_, _, scriptPath, bashProfilePath, bashrcPath, err := bashIntegrationPaths()
	if err != nil {
		return err
	}
	scriptStatus, profileStatus, rcStatus, summary := computeInstallStatuses(scriptPath, bashProfilePath, bashrcPath, true)
	printInstallReport(summary, scriptPath, scriptStatus, bashProfilePath, profileStatus, bashrcPath, rcStatus)
	return nil
}

func uninstallBashIntegrationDryRun() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	bashProfilePath := filepath.Join(home, ".bash_profile")
	bashrcPath := filepath.Join(home, ".bashrc")

	if fullyUninstalled(home) {
		fmt.Println("wrk bash integration: already uninstalled")
		fmt.Printf("bash_profile: %s (marker absent)\n", bashProfilePath)
		fmt.Printf("bashrc: %s (marker absent)\n", bashrcPath)
		fmt.Println("no changes needed")
		fmt.Println()
		return nil
	}

	fmt.Println("dry-run: would remove marker block from ~/.bash_profile")
	fmt.Println("dry-run: would remove marker block from ~/.bashrc")
	fmt.Println()
	fmt.Print(wrkMarkerBlock)
	fmt.Println()
	return nil
}

func statusBashIntegration() (exitCode int) {
	_, wrkHome, scriptPath, bashProfilePath, bashrcPath, err := bashIntegrationPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	scriptExists := scriptPresent(scriptPath)
	profileMarker := markerPresent(bashProfilePath)
	bashrcMarker := markerPresent(bashrcPath)

	var state string
	exitCode = 1
	switch {
	case scriptExists && profileMarker && bashrcMarker:
		state = "installed"
		exitCode = 0
	case !scriptExists && !profileMarker && !bashrcMarker:
		state = "not installed"
	default:
		state = "partial"
	}

	fmt.Printf("bash integration: %s\n", state)
	if scriptExists {
		fmt.Printf("script: %s (present)\n", scriptPath)
	} else {
		fmt.Printf("script: %s (absent)\n", scriptPath)
	}
	if profileMarker {
		fmt.Printf("bash_profile: %s (marker present)\n", bashProfilePath)
	} else {
		fmt.Printf("bash_profile: %s (marker absent)\n", bashProfilePath)
	}
	if bashrcMarker {
		fmt.Printf("bashrc: %s (marker present)\n", bashrcPath)
	} else {
		fmt.Printf("bashrc: %s (marker absent)\n", bashrcPath)
	}
	fmt.Println()
	_ = wrkHome
	return exitCode
}

func installBashIntegration() error {
	_, _, scriptPath, bashProfilePath, bashrcPath, err := bashIntegrationPaths()
	if err != nil {
		return err
	}

	// Compute statuses from pre-write filesystem state.
	scriptStatus, profileStatus, rcStatus, summary := computeInstallStatuses(scriptPath, bashProfilePath, bashrcPath, false)

	integrationDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(integrationDir, 0o755); err != nil {
		return err
	}
	// Ensure bash.sh matches the current embedded script so upgrades pick up
	// the wrk() auto-cd wrapper (rewrite when missing or content differs).
	want := []byte(bashIntegrationScript)
	existing, readErr := os.ReadFile(scriptPath)
	if readErr != nil || string(existing) != string(want) {
		if err := os.WriteFile(scriptPath, want, 0o644); err != nil {
			return err
		}
	}

	for _, profilePath := range []string{bashProfilePath, bashrcPath} {
		if err := appendMarkerToProfile(profilePath); err != nil {
			return err
		}
	}

	printInstallReport(summary, scriptPath, scriptStatus, bashProfilePath, profileStatus, bashrcPath, rcStatus)
	return nil
}

// computeInstallStatuses inspects script and profile markers and returns
// per-component statuses plus summary. dryRun selects would-* vocabulary.
func computeInstallStatuses(scriptPath, bashProfilePath, bashrcPath string, dryRun bool) (scriptStatus, profileStatus, rcStatus, summary string) {
	existing, readErr := os.ReadFile(scriptPath)
	scriptMissing := readErr != nil
	scriptDiffers := !scriptMissing && string(existing) != bashIntegrationScript

	profileHas := markerPresent(bashProfilePath)
	rcHas := markerPresent(bashrcPath)

	if dryRun {
		switch {
		case scriptMissing:
			scriptStatus = "would install"
		case scriptDiffers:
			scriptStatus = "would update"
		default:
			scriptStatus = "is up to date"
		}
		if profileHas {
			profileStatus = "is up to date"
		} else {
			profileStatus = "would install"
		}
		if rcHas {
			rcStatus = "is up to date"
		} else {
			rcStatus = "would install"
		}
		switch {
		case scriptMissing:
			summary = "would install"
		case scriptDiffers || !profileHas || !rcHas:
			summary = "would update"
		default:
			summary = "is up to date"
		}
		return scriptStatus, profileStatus, rcStatus, summary
	}

	switch {
	case scriptMissing:
		scriptStatus = "installed"
	case scriptDiffers:
		scriptStatus = "updated"
	default:
		scriptStatus = "is up to date"
	}
	if profileHas {
		profileStatus = "is up to date"
	} else {
		profileStatus = "installed"
	}
	if rcHas {
		rcStatus = "is up to date"
	} else {
		rcStatus = "installed"
	}
	switch {
	case scriptMissing:
		summary = "installed"
	case scriptDiffers || !profileHas || !rcHas:
		summary = "updated"
	default:
		summary = "is up to date"
	}
	return scriptStatus, profileStatus, rcStatus, summary
}

func printInstallReport(summary, scriptPath, scriptStatus, bashProfilePath, profileStatus, bashrcPath, rcStatus string) {
	fmt.Printf("bash integration: %s\n", summary)
	fmt.Printf("script: %s (%s)\n", scriptPath, scriptStatus)
	fmt.Printf("bash_profile: %s (marker %s)\n", bashProfilePath, profileStatus)
	fmt.Printf("bashrc: %s (marker %s)\n", bashrcPath, rcStatus)
	fmt.Println()
}

func appendMarkerToProfile(profilePath string) error {
	profile, err := os.ReadFile(profilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(profile)
	if strings.Contains(content, wrkMarkerBegin) {
		return nil
	}

	var builder strings.Builder
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		builder.WriteString(content)
		builder.WriteString("\n")
	} else {
		builder.WriteString(content)
	}
	builder.WriteString(wrkMarkerBlock)
	return os.WriteFile(profilePath, []byte(builder.String()), 0o644)
}

func uninstallBashIntegration() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	for _, profilePath := range []string{
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bashrc"),
	} {
		if err := stripMarkerFromProfile(profilePath); err != nil {
			return err
		}
	}
	return nil
}

func stripMarkerFromProfile(profilePath string) error {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == wrkMarkerBegin {
			inBlock = true
			continue
		}
		if trimmed == wrkMarkerEnd {
			inBlock = false
			continue
		}
		if inBlock {
			continue
		}
		out = append(out, line)
	}

	newContent := strings.Join(out, "\n")
	if len(data) > 0 && strings.HasSuffix(string(data), "\n") && newContent != "" {
		newContent += "\n"
	}
	return os.WriteFile(profilePath, []byte(newContent), 0o644)
}

func runBashComplete(req *CompletionRequest) error {
	if req == nil {
		return fmt.Errorf("wrk: --complete requires words and cword after --")
	}
	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	candidates := Complete(wrkHome, *req)
	if len(candidates) == 0 {
		return nil
	}
	for _, c := range candidates {
		fmt.Println(c)
	}
	fmt.Println()
	return nil
}

// isPathLike reports whether cur should yield to bash default filename
// completion rather than wrk basenames/flags. Path-like prefixes: /, ./, ../.
func isPathLike(cur string) bool {
	return strings.HasPrefix(cur, "/") ||
		strings.HasPrefix(cur, "./") ||
		strings.HasPrefix(cur, "../")
}

// Complete returns bash tab-completion candidates for the given request.
func Complete(wrkHome string, req CompletionRequest) []string {
	if req.CWord < 0 || req.CWord >= len(req.Words) {
		return nil
	}
	cur := req.Words[req.CWord]
	if isPathLike(cur) {
		return nil
	}

	kind, prefix := completionContext(req.Words, req.CWord)
	switch kind {
	case "flags":
		return filterFlags(prefix)
	case "basenames":
		candidates, err := listBasenameCandidates(wrkHome, prefix)
		if err != nil {
			return nil
		}
		return candidates
	default:
		return nil
	}
}

func completionContext(words []string, cword int) (kind, prefix string) {
	if cword < 0 || cword >= len(words) {
		return "none", ""
	}
	cur := words[cword]

	if strings.HasPrefix(cur, "-") {
		return "flags", cur
	}

	if cword > 0 {
		switch words[cword-1] {
		case "--bring", "--where", "--add", "--rm", "--cd", "-l", "--list", "--status":
			return "basenames", cur
		case "-t", "--task", "--set-task":
			return "none", ""
		}
	}

	if cword == 1 && !strings.HasPrefix(cur, "-") {
		return "basenames", cur
	}

	if cword == 2 && len(words) > 1 && !strings.HasPrefix(words[1], "-") {
		return "basenames", cur
	}

	return "none", ""
}

func filterFlags(prefix string) []string {
	var out []string
	for _, flag := range wrkCompletionFlags {
		if strings.HasPrefix(flag, prefix) {
			out = append(out, flag)
		}
	}
	return out
}

func listBasenameCandidates(wrkHome, prefix string) ([]string, error) {
	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var basenames []string
	for _, p := range paths {
		base := filepath.Base(p)
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		if prefix == "" || strings.HasPrefix(base, prefix) {
			basenames = append(basenames, base)
		}
	}
	sort.Strings(basenames)
	return basenames, nil
}
