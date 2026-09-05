package provider

import (
	"io"
	"strings"
)

// Brew wraps Homebrew. Formulae only; never sudo.
type Brew struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (b *Brew) Name() string { return "brew" }

func (b *Brew) Installed(pkg string) (bool, error) {
	if _, err := query("brew", "list", pkg); err != nil {
		return false, nil
	}
	return true, nil
}

func (b *Brew) Install(pkg string) error {
	return run(b.stdin, b.stdout, b.stderr, "brew", "install", pkg)
}

func (b *Brew) Remove(pkg string) error {
	return run(b.stdin, b.stdout, b.stderr, "brew", "uninstall", pkg)
}

func (b *Brew) Upgrade(pkg string) error {
	return run(b.stdin, b.stdout, b.stderr, "brew", "upgrade", pkg)
}

func (b *Brew) UpgradeAll() error {
	return run(b.stdin, b.stdout, b.stderr, "brew", "upgrade")
}

func (b *Brew) List() ([]string, error) {
	out, err := query("brew", "list", "--formula")
	if err != nil {
		return nil, err
	}
	return lines(out), nil
}

func (b *Brew) Search(term string) ([]string, error) {
	out, err := query("brew", "search", term)
	if err != nil {
		return nil, err
	}
	var matches []string
	seen := map[string]bool{}
	for _, l := range lines(out) {
		if strings.HasPrefix(l, "==>") || strings.HasPrefix(l, "No formulae or casks") {
			continue
		}
		for _, tok := range strings.Fields(l) {
			if !seen[tok] {
				seen[tok] = true
				matches = append(matches, tok)
			}
		}
	}
	return matches, nil
}
