package wrkcli

import "fmt"

// WebServeOptions configures the local web UI started by `wrk --web`.
type WebServeOptions struct {
	WrkHome string
	Port    int
	Dev     bool
}

// webServeFunc starts the local web UI. Registered from cmd/wrk to avoid an
// import cycle (wrkserver imports wrkcli).
type webServeFunc func(opts WebServeOptions) error

var webServeImpl webServeFunc

// RegisterWebServe sets the implementation used by `wrk --web`.
// cmd/wrk calls this from init with wrkcli/web.Serve.
func RegisterWebServe(fn webServeFunc) {
	webServeImpl = fn
}

func runWeb(opts WebServeOptions) error {
	if webServeImpl == nil {
		return fmt.Errorf("wrk: --web is not available in this build")
	}
	return webServeImpl(opts)
}
