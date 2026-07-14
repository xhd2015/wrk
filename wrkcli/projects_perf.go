package wrkcli

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

const envProjectsPerfLog = "WRK_PROJECTS_PERF_LOG"

type perfEvent struct {
	Event      string `json:"event"`
	Project    string `json:"project,omitempty"`
	Phase      string `json:"phase,omitempty"`
	Worktree   string `json:"worktree,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty"`
	Count      int    `json:"count,omitempty"`
	TotalMS    int64  `json:"total_ms,omitempty"`
}

type projectsPerf struct {
	file         *os.File
	mu           sync.Mutex
	project      string
	projectStart time.Time
	runStart     time.Time
}

var currentProjectsPerf *projectsPerf

func beginProjectsPerfRun() func() {
	path := os.Getenv(envProjectsPerfLog)
	if path == "" {
		return func() {}
	}
	f, err := os.Create(path)
	if err != nil {
		return func() {}
	}
	p := &projectsPerf{file: f, runStart: time.Now()}
	currentProjectsPerf = p
	p.emit(perfEvent{Event: "run_start"})
	return func() {
		p.emit(perfEvent{Event: "run_end", TotalMS: time.Since(p.runStart).Milliseconds()})
		f.Close()
		currentProjectsPerf = nil
	}
}

func beginProjectPerf(project string) func() {
	p := currentProjectsPerf
	if p == nil {
		return func() {}
	}
	p.mu.Lock()
	p.project = project
	p.projectStart = time.Now()
	p.emitLocked(perfEvent{Event: "project_start", Project: project})
	p.mu.Unlock()
	return func() {
		p.emit(perfEvent{
			Event:   "project_end",
			Project: project,
			TotalMS: time.Since(p.projectStart).Milliseconds(),
		})
	}
}

func recordProjectsPerfPhase(project, phase string, d time.Duration) {
	p := currentProjectsPerf
	if p == nil {
		return
	}
	p.emit(perfEvent{
		Event:      "phase",
		Project:    project,
		Phase:      phase,
		DurationMS: d.Milliseconds(),
	})
}

func recordProjectsPerfWorktree(project, wtPath string, d time.Duration) {
	p := currentProjectsPerf
	if p == nil {
		return
	}
	p.emit(perfEvent{
		Event:      "worktree_status",
		Project:    project,
		Worktree:   wtPath,
		DurationMS: d.Milliseconds(),
	})
}

func recordProjectsPerfAggregate(project, phase string, count int, d time.Duration) {
	p := currentProjectsPerf
	if p == nil {
		return
	}
	p.emit(perfEvent{
		Event:      "phase_total",
		Project:    project,
		Phase:      phase,
		Count:      count,
		DurationMS: d.Milliseconds(),
	})
}

func projectsPerfTimed(project, phase string, fn func() error) error {
	p := currentProjectsPerf
	if p == nil {
		return fn()
	}
	start := time.Now()
	err := fn()
	recordProjectsPerfPhase(project, phase, time.Since(start))
	return err
}

func projectsPerfTimedValue[T any](project, phase string, fn func() (T, error)) (T, error) {
	p := currentProjectsPerf
	if p == nil {
		return fn()
	}
	start := time.Now()
	v, err := fn()
	recordProjectsPerfPhase(project, phase, time.Since(start))
	return v, err
}

func (p *projectsPerf) emit(ev perfEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.emitLocked(ev)
}

func (p *projectsPerf) emitLocked(ev perfEvent) {
	data, err := json.Marshal(ev)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = p.file.Write(data)
}