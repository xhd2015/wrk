package main

import (
	"errors"
	"fmt"
	"os"

	// teapre before wrkcli: skip bubbletea/lipgloss OSC background hang on dumb PTYs.
	_ "github.com/xhd2015/wrk/wrkcli/teapre"
	"github.com/xhd2015/wrk/wrkcli"
	"github.com/xhd2015/wrk/wrkcli/web"
)

func init() {
	// Register outside wrkcli to avoid import cycle: wrkserver imports wrkcli.
	wrkcli.RegisterWebServe(func(opts wrkcli.WebServeOptions) error {
		return web.Serve(web.Options{
			WrkHome: opts.WrkHome,
			Port:    opts.Port,
			Dev:     opts.Dev,
		})
	})
}

func main() {
	if err := run(); err != nil {
		var exitErr wrkcli.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		// Color Error: (red) / warning: (yellow) prefixes when --color or TTY
		// and NO_COLOR is unset. Prefix-only; body stays plain.
		msg := err.Error()
		msg = wrkcli.FormatStderrError(msg)
		msg = wrkcli.FormatStderrWarning(msg)
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
}

func run() error {
	return wrkcli.Run(os.Args[1:])
}
