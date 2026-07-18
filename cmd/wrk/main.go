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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	return wrkcli.Run(os.Args[1:])
}
