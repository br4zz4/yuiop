// Package platform resolves the effective package manager for this machine.
package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var supported = []string{"brew", "apt", "pacman"}

// Validate returns the platform name if it is supported.
func Validate(name string) (string, error) {
	for _, p := range supported {
		if name == p {
			return name, nil
		}
	}
	return "", fmt.Errorf("unsupported platform %q (supported: %s)", name, strings.Join(supported, ", "))
}

// Resolve applies the precedence: CLI > env > config > auto-detect.
func Resolve(cli, env, config string) (string, error) {
	switch {
	case cli != "":
		return Validate(cli)
	case env != "":
		return Validate(env)
	case config != "":
		return Validate(config)
	}
	return Detect()
}

// Detect chooses the default provider from the host OS and available binaries.
func Detect() (string, error) {
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err != nil {
			return "", fmt.Errorf("macOS detected but brew is not on PATH")
		}
		return "brew", nil
	}
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/usr/bin/apt-get"); err == nil {
			return "apt", nil
		}
		if _, err := os.Stat("/usr/bin/pacman"); err == nil {
			return "pacman", nil
		}
	}
	return "", fmt.Errorf("no supported platform detected (macOS, Debian, Arch)")
}
