package provider

import (
	"io"
	"strings"
)

// Apt wraps Debian's apt-get. Uses sudo when not root.
type Apt struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (a *Apt) Name() string { return "apt" }

func (a *Apt) Installed(pkg string) (bool, error) {
	if _, err := query("dpkg", "-s", pkg); err != nil {
		return false, nil
	}
	return true, nil
}

func (a *Apt) Install(pkg string) error {
	args := sudo("apt-get", "install", "-y", pkg)
	return run(a.stdin, a.stdout, a.stderr, args[0], args[1:]...)
}

func (a *Apt) Remove(pkg string) error {
	args := sudo("apt-get", "remove", "-y", pkg)
	return run(a.stdin, a.stdout, a.stderr, args[0], args[1:]...)
}

func (a *Apt) Upgrade(pkg string) error {
	args := sudo("apt-get", "install", "--only-upgrade", "-y", pkg)
	return run(a.stdin, a.stdout, a.stderr, args[0], args[1:]...)
}

func (a *Apt) UpgradeAll() error {
	args := sudo("apt-get", "upgrade", "-y")
	return run(a.stdin, a.stdout, a.stderr, args[0], args[1:]...)
}

func (a *Apt) List() ([]string, error) {
	out, err := query("dpkg-query", "-W", "-f=${Package}\n")
	if err != nil {
		return nil, err
	}
	return lines(out), nil
}

func (a *Apt) Search(term string) ([]string, error) {
	out, err := query("apt-cache", "search", term)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, l := range lines(out) {
		if i := strings.Index(l, " "); i > 0 {
			l = l[:i]
		}
		names = append(names, l)
	}
	return names, nil
}
