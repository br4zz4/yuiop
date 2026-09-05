// Package cli implements the yuiop command-line interface.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/br4zz4/yuiop"
	"github.com/br4zz4/yuiop/internal/config"
	"github.com/br4zz4/yuiop/internal/platform"
	"github.com/br4zz4/yuiop/internal/provider"
	"github.com/br4zz4/yuiop/internal/resolve"
)

// Exit codes, documented in docs/CONTRACT.md.
const (
	ExitOK       = 0
	ExitFailure  = 1
	ExitUsage    = 2
	ExitNotFound = 3
)

const Version = "0.1.0"

type options struct {
	jsonOut      bool
	platformFlag string
	configPath   string
}

type app struct {
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	jsonOut  bool
	opts     options
	cfg      *config.Config
	provider provider.Provider
	table    *resolve.Table
}

// Run parses arguments and dispatches to a command, returning the exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var opts options
	rest, err := parseFlags(args, &opts)
	if err != nil {
		fmt.Fprintf(stderr, "yuiop: %v\n", err)
		return ExitUsage
	}
	if len(rest) == 0 {
		usage(stderr)
		return ExitUsage
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		fmt.Fprintf(stderr, "yuiop: %v\n", err)
		return ExitFailure
	}

	pl, err := platform.Resolve(opts.platformFlag, os.Getenv("YUIOP_PLATFORM"), cfg.Platform)
	if err != nil {
		fmt.Fprintf(stderr, "yuiop: %v\n", err)
		return ExitFailure
	}

	provStdout := stdout
	if opts.jsonOut {
		provStdout = io.Discard
	}
	prov, err := provider.For(pl, stdin, provStdout, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "yuiop: %v\n", err)
		return ExitFailure
	}

	table, err := resolve.Load(yuiop.PackagesYAML)
	if err != nil {
		fmt.Fprintf(stderr, "yuiop: %v\n", err)
		return ExitFailure
	}

	a := &app{stdin: stdin, stdout: stdout, stderr: stderr, jsonOut: opts.jsonOut, opts: opts, cfg: cfg, provider: prov, table: table}
	return a.dispatch(rest[0], rest[1:])
}

func (a *app) dispatch(cmd string, args []string) int {
	switch cmd {
	case "install":
		return a.requireOne(cmd, args, a.install)
	case "remove":
		return a.requireOne(cmd, args, a.remove)
	case "upgrade":
		return a.upgrade(args)
	case "list":
		return a.list()
	case "search":
		return a.requireOne(cmd, args, a.search)
	case "info":
		return a.requireOne(cmd, args, a.info)
	case "status":
		return a.requireOne(cmd, args, a.status)
	case "platform":
		return a.platform(args)
	case "version":
		fmt.Fprintf(a.stdout, "yuiop %s\n", Version)
		return ExitOK
	}
	usage(a.stderr)
	return ExitUsage
}

func (a *app) requireOne(cmd string, args []string, fn func(string) int) int {
	if len(args) != 1 {
		fmt.Fprintf(a.stderr, "yuiop: %s requires exactly one package name\n", cmd)
		return ExitUsage
	}
	return fn(args[0])
}

func (a *app) install(canonical string) int {
	pkg, ok := a.table.Resolve(canonical, a.provider.Name())
	if !ok {
		fmt.Fprintf(a.stderr, "yuiop: no knowledge of package '%s'\n", canonical)
		return ExitNotFound
	}
	installed, err := a.provider.Installed(pkg)
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if installed {
		a.emit(map[string]any{"status": "already-installed", "package": canonical, "resolved": pkg})
		return ExitOK
	}
	if err := a.provider.Install(pkg); err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %s install failed: %v\n", pkg, err)
		return ExitFailure
	}
	a.emit(map[string]any{"status": "installed", "package": canonical, "resolved": pkg})
	return ExitOK
}

func (a *app) remove(canonical string) int {
	pkg, ok := a.table.Resolve(canonical, a.provider.Name())
	if !ok {
		fmt.Fprintf(a.stderr, "yuiop: no knowledge of package '%s'\n", canonical)
		return ExitNotFound
	}
	installed, err := a.provider.Installed(pkg)
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if !installed {
		a.emit(map[string]any{"status": "not-installed", "package": canonical})
		return ExitOK
	}
	if err := a.provider.Remove(pkg); err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %s remove failed: %v\n", pkg, err)
		return ExitFailure
	}
	a.emit(map[string]any{"status": "removed", "package": canonical, "resolved": pkg})
	return ExitOK
}

func (a *app) upgrade(args []string) int {
	for _, x := range args {
		if x == "--all" {
			if err := a.provider.UpgradeAll(); err != nil {
				fmt.Fprintf(a.stderr, "yuiop: upgrade failed: %v\n", err)
				return ExitFailure
			}
			a.emit(map[string]any{"status": "upgraded"})
			return ExitOK
		}
	}
	if len(args) == 0 {
		if err := a.provider.UpgradeAll(); err != nil {
			fmt.Fprintf(a.stderr, "yuiop: upgrade failed: %v\n", err)
			return ExitFailure
		}
		a.emit(map[string]any{"status": "upgraded"})
		return ExitOK
	}
	if len(args) == 1 {
		return a.requireOne("upgrade", args, a.upgradeOne)
	}
	fmt.Fprintf(a.stderr, "yuiop: upgrade takes at most one package name or --all\n")
	return ExitUsage
}

func (a *app) upgradeOne(canonical string) int {
	pkg, ok := a.table.Resolve(canonical, a.provider.Name())
	if !ok {
		fmt.Fprintf(a.stderr, "yuiop: no knowledge of package '%s'\n", canonical)
		return ExitNotFound
	}
	if err := a.provider.Upgrade(pkg); err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %s upgrade failed: %v\n", pkg, err)
		return ExitFailure
	}
	a.emit(map[string]any{"status": "upgraded", "package": canonical, "resolved": pkg})
	return ExitOK
}

func (a *app) list() int {
	pkgs, err := a.provider.List()
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if a.jsonOut {
		a.emit(map[string]any{"platform": a.provider.Name(), "packages": pkgs})
		return ExitOK
	}
	for _, p := range pkgs {
		fmt.Fprintln(a.stdout, p)
	}
	return ExitOK
}

func (a *app) search(term string) int {
	matches, err := a.provider.Search(term)
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if len(matches) == 0 {
		return ExitNotFound
	}
	if a.jsonOut {
		a.emit(map[string]any{"platform": a.provider.Name(), "term": term, "matches": matches})
		return ExitOK
	}
	for _, m := range matches {
		fmt.Fprintln(a.stdout, m)
	}
	return ExitOK
}

func (a *app) info(canonical string) int {
	pkg, ok := a.table.Resolve(canonical, a.provider.Name())
	if !ok {
		fmt.Fprintf(a.stderr, "yuiop: no knowledge of package '%s'\n", canonical)
		return ExitNotFound
	}
	installed, err := a.provider.Installed(pkg)
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if a.jsonOut {
		a.emit(map[string]any{"platform": a.provider.Name(), "package": canonical, "resolved": pkg, "installed": installed})
		return ExitOK
	}
	fmt.Fprintf(a.stdout, "%s -> %s (%s)\n", canonical, pkg, installedStatus(installed))
	return ExitOK
}

func (a *app) status(canonical string) int {
	pkg, ok := a.table.Resolve(canonical, a.provider.Name())
	if !ok {
		fmt.Fprintf(a.stderr, "yuiop: no knowledge of package '%s'\n", canonical)
		return ExitNotFound
	}
	installed, err := a.provider.Installed(pkg)
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if a.jsonOut {
		a.emit(map[string]any{"platform": a.provider.Name(), "package": canonical, "installed": installed})
		return ExitOK
	}
	fmt.Fprintf(a.stdout, "%s: %s\n", canonical, installedStatus(installed))
	return ExitOK
}

func (a *app) platform(args []string) int {
	if len(args) == 0 {
		if a.jsonOut {
			a.emit(map[string]string{"platform": a.provider.Name()})
		} else {
			fmt.Fprintln(a.stdout, a.provider.Name())
		}
		return ExitOK
	}
	name, err := platform.Validate(args[0])
	if err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitUsage
	}
	if err := config.SetPlatform(name, a.opts.configPath); err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
		return ExitFailure
	}
	if a.jsonOut {
		a.emit(map[string]string{"platform": name})
	}
	return ExitOK
}

func (a *app) emit(v any) {
	if !a.jsonOut {
		return
	}
	if err := json.NewEncoder(a.stdout).Encode(v); err != nil {
		fmt.Fprintf(a.stderr, "yuiop: %v\n", err)
	}
}

func installedStatus(b bool) string {
	if b {
		return "installed"
	}
	return "not installed"
}

func parseFlags(args []string, opts *options) ([]string, error) {
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			opts.jsonOut = true
		case a == "--platform" || a == "--config":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag %s requires a value", a)
			}
			i++
			if a == "--platform" {
				opts.platformFlag = args[i]
			} else {
				opts.configPath = args[i]
			}
		case strings.HasPrefix(a, "--platform="):
			opts.platformFlag = strings.TrimPrefix(a, "--platform=")
		case strings.HasPrefix(a, "--config="):
			opts.configPath = strings.TrimPrefix(a, "--config=")
		case strings.HasPrefix(a, "-"):
			return nil, fmt.Errorf("unknown flag %q", a)
		default:
			rest = append(rest, a)
		}
	}
	return rest, nil
}

func usage(w io.Writer) {
	fmt.Fprint(w, `yuiop — the last package manager you need to know

Usage:
  yuiop [--json] [--platform <brew|apt|pacman>] [--config <path>] <command> [args]

Commands:
  install <pkg>     Install a system package via the platform manager
  remove <pkg>      Remove a package
  list              List installed packages
  search <term>     Search packages
  upgrade [<pkg>]   Upgrade one package, or all when no package is given
  info <pkg>        Show the resolved name and install status
  status <pkg>      Report whether a package is installed (machine contract)
  platform [<name>] Show the effective platform, or persist an override
  version           Print the yuiop version
`)
}
