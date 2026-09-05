// Package provider defines the package-manager abstraction and its
// implementations for brew, apt and pacman.
package provider

import (
	"fmt"
	"io"
)

// Provider is a thin adapter over a system package manager. Operations that
// change state pass stdin/stdout/stderr through so prompts (e.g. sudo) work.
type Provider interface {
	Name() string
	Installed(pkg string) (bool, error)
	Install(pkg string) error
	Remove(pkg string) error
	Upgrade(pkg string) error
	UpgradeAll() error
	List() ([]string, error)
	Search(term string) ([]string, error)
}

// For returns the provider for the given platform name.
func For(name string, stdin io.Reader, stdout, stderr io.Writer) (Provider, error) {
	switch name {
	case "brew":
		return &Brew{stdin: stdin, stdout: stdout, stderr: stderr}, nil
	case "apt":
		return &Apt{stdin: stdin, stdout: stdout, stderr: stderr}, nil
	case "pacman":
		return &Pacman{stdin: stdin, stdout: stdout, stderr: stderr}, nil
	}
	return nil, fmt.Errorf("no provider for %q", name)
}
