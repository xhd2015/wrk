// Package web serves the wrk React SPA (wrk-react) and mounts wrkserver.
//
// Kept as a separate package so wrkcli can avoid importing wrkserver (which
// imports wrkcli), preventing an import cycle. cmd/wrk registers Serve via
// wrkcli.RegisterWebServe.
package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/wrk/wrkcli/wrkserver"
)

// distFS holds the production SPA build (copied from wrk-react/dist by script/build-frontend).
//
//go:embed all:dist
var distFS embed.FS

// Options configures the local web UI server.
type Options struct {
	WrkHome string
	Port    int
	// Dev proxies non-API traffic to the Vite server under wrk-react/.
	Dev bool
}

// Serve binds 127.0.0.1:port (or auto-picks from 8080 when port==0), prints the
// listen URL to stdout, serves the React SPA (or Vite in Dev) and mounts
// wrkserver at /api/wrk. Blocks until SIGINT/SIGTERM or server error.
func Serve(opts Options) error {
	ln, actualPort, err := listenLocal(opts.Port)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	wrkserver.New(wrkserver.Options{WrkHome: opts.WrkHome}).Register(mux, "/api/wrk")

	var viteDone <-chan struct{}
	if opts.Dev {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		vitePort, done, err := ensureFrontendDevServer(ctx)
		if err != nil {
			return err
		}
		viteDone = done
		if err := proxyDev(mux, vitePort); err != nil {
			cancel()
			return err
		}
	} else {
		if err := mountStaticSPA(mux); err != nil {
			return err
		}
	}

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Print before Serve so probes can discover the URL while the process runs.
	fmt.Printf("http://127.0.0.1:%d/\n", actualPort)

	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if err == nil || err == http.ErrServerClosed {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		<-errCh
		if viteDone != nil {
			// Wait for Vite child to exit after parent cancel via Shutdown path:
			// ensureFrontendDevServer is cancelled when Serve returns via defer cancel
			// only when Dev was true — cancel is deferred at start of Dev block.
			// Here we just wait a moment if channel already closed.
			select {
			case <-viteDone:
			case <-time.After(2 * time.Second):
			}
		}
		return nil
	}
}

func listenLocal(port int) (net.Listener, int, error) {
	if port < 0 {
		return nil, 0, fmt.Errorf("wrk: invalid --port %d", port)
	}
	if port > 0 {
		ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			return nil, 0, fmt.Errorf("wrk: listen 127.0.0.1:%d: %w", port, err)
		}
		return ln, port, nil
	}
	for p := 8080; p < 8080+200; p++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p)))
		if err == nil {
			return ln, p, nil
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, 0, fmt.Errorf("wrk: listen 127.0.0.1: %w", err)
	}
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		_ = ln.Close()
		return nil, 0, fmt.Errorf("wrk: listen: unexpected address type %T", ln.Addr())
	}
	return ln, addr.Port, nil
}

func mountStaticSPA(mux *http.ServeMux) error {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return fmt.Errorf("wrk: embed spa dist: %w", err)
	}
	fileServer := http.FileServer(http.FS(sub))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := r.URL.Path
		if path == "/" {
			serveIndex(w, r, sub)
			return
		}
		// Try static asset first.
		clean := strings.TrimPrefix(path, "/")
		if clean != "" {
			if f, err := sub.Open(clean); err == nil {
				_ = f.Close()
				// Set MIME for common assets when FileServer is picky on embed.
				if ext := filepath.Ext(clean); ext != "" {
					if mt := mime.TypeByExtension(ext); mt != "" {
						w.Header().Set("Content-Type", mt)
					}
				}
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// SPA client-route fallback (e.g. /mockup/repo-view).
		serveIndex(w, r, sub)
	})
	return nil
}

func serveIndex(w http.ResponseWriter, r *http.Request, sub fs.FS) {
	data, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		http.Error(w, "spa index unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(data)
}

func proxyDev(mux *http.ServeMux, vitePort int) error {
	targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", vitePort))
	if err != nil {
		return fmt.Errorf("invalid vite proxy target: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// API is registered on more specific paths; this catches the rest.
		r.Host = targetURL.Host
		proxy.ServeHTTP(w, r)
	})
	return nil
}

func ensureFrontendDevServer(ctx context.Context) (int, <-chan struct{}, error) {
	reactDir, err := findReactDir()
	if err != nil {
		return 0, nil, err
	}
	if _, err := exec.LookPath("bun"); err != nil {
		return 0, nil, fmt.Errorf("wrk: --dev requires bun on PATH: %w", err)
	}
	if _, err := os.Stat(filepath.Join(reactDir, "node_modules")); err != nil {
		install := exec.CommandContext(ctx, "bun", "install")
		install.Dir = reactDir
		install.Stdout = os.Stderr
		install.Stderr = os.Stderr
		if err := install.Run(); err != nil {
			return 0, nil, fmt.Errorf("wrk: bun install in %s: %w", reactDir, err)
		}
	}

	vitePort, err := findFreeLoopbackPort(5173, 100)
	if err != nil {
		return 0, nil, err
	}

	fmt.Fprintf(os.Stderr, "Starting frontend dev server on port %d (%s)...\n", vitePort, reactDir)
	cmd := exec.CommandContext(ctx, "bun", "run", "dev", "--", "--port", strconv.Itoa(vitePort), "--strictPort")
	cmd.Dir = reactDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return 0, nil, fmt.Errorf("wrk: start vite: %w", err)
	}

	childExited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(childExited)
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			<-childExited
		case <-childExited:
		}
	}()

	fmt.Fprintf(os.Stderr, "Waiting for frontend server on port %d...", vitePort)
	for i := 0; i < 60; i++ {
		select {
		case <-childExited:
			fmt.Fprintln(os.Stderr)
			return 0, nil, fmt.Errorf("wrk: frontend dev server exited before ready")
		default:
		}
		if portOpen(vitePort) {
			fmt.Fprintln(os.Stderr, " Ready!")
			return vitePort, done, nil
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Fprint(os.Stderr, ".")
	}
	fmt.Fprintln(os.Stderr)
	return 0, nil, fmt.Errorf("wrk: frontend dev server failed to start within timeout")
}

func findReactDir() (string, error) {
	if env := strings.TrimSpace(os.Getenv("WRK_REACT_DIR")); env != "" {
		if st, err := os.Stat(filepath.Join(env, "package.json")); err == nil && !st.IsDir() {
			return env, nil
		}
		return "", fmt.Errorf("wrk: WRK_REACT_DIR=%s has no package.json", env)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		candidate := filepath.Join(dir, "wrk-react")
		pkg := filepath.Join(candidate, "package.json")
		if st, err := os.Stat(pkg); err == nil && !st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("wrk: could not find wrk-react/ package (set WRK_REACT_DIR or run from the wrk repo)")
}

func findFreeLoopbackPort(start, maxAttempts int) (int, error) {
	for i := 0; i < maxAttempts; i++ {
		p := start + i
		if !portOpen(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("wrk: no free vite port in [%d, %d)", start, start+maxAttempts)
}

func portOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

