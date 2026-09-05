package provider

import (
	"io"
	"strings"
)

// Pacman wraps Arch's pacman. Uses sudo when not root.
type Pacman struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (p *Pacman) Name() string { return "pacman" }

func (p *Pacman) Installed(pkg string) (bool, error) {
	if _, err := query("pacman", "-Q", pkg); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *Pacman) Install(pkg string) error {
	args := sudo("pacman", "-S", "--noconfirm", pkg)
	return run(p.stdin, p.stdout, p.stderr, args[0], args[1:]...)
}

func (p *Pacman) Remove(pkg string) error {
	args := sudo("pacman", "-R", "--noconfirm", pkg)
	return run(p.stdin, p.stdout, p.stderr, args[0], args[1:]...)
}

func (p *Pacman) Upgrade(pkg string) error {
	args := sudo("pacman", "-S", "--noconfirm", pkg)
	return run(p.stdin, p.stdout, p.stderr, args[0], args[1:]...)
}

func (p *Pacman) UpgradeAll() error {
	args := sudo("pacman", "-Su", "--noconfirm")
	return run(p.stdin, p.stdout, p.stderr, args[0], args[1:]...)
}

func (p *Pacman) List() ([]string, error) {
	out, err := query("pacman", "-Qq")
	if err != nil {
		return nil, err
	}
	return lines(out), nil
}

func (p *Pacman) Search(term string) ([]string, error) {
	out, err := query("pacman", "-Ss", term)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, l := range lines(out) {
		name := l
		if i := strings.Index(l, "/"); i > 0 {
			name = l[:i]
		}
		names = append(names, name)
	}
	return names, nil
}
