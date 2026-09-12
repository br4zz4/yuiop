package provider

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Aur wraps an AUR helper (yay or paru) for Arch Linux user-repository
// packages. AUR packages are PKGBUILD recipes — they need a helper, not pacman.
type Aur struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	helper string // resolved helper binary ("yay" or "paru")
}

// Name returns the provider name.
func (a *Aur) Name() string { return "aur" }

// helperFor returns the AUR helper binary, preferring yay then paru.
func (a *Aur) helperFor() (string, error) {
	for _, h := range []string{"yay", "paru"} {
		if _, err := exec.LookPath(h); err == nil {
			return h, nil
		}
	}
	return "", fmt.Errorf(
		"no AUR helper found (need yay or paru) — install it with:\n  sudo pacman -S --needed git base-devel\n  git clone https://aur.archlinux.org/yay.git && cd yay && makepkg -si",
	)
}

// Installed reports whether a package is installed (via the helper's query).
func (a *Aur) Installed(pkg string) (bool, error) {
	h, err := a.helperFor()
	if err != nil {
		return false, err
	}
	if _, err := query(h, "-Q", pkg); err != nil {
		return false, nil
	}
	return true, nil
}

// Install installs a package from the AUR.
func (a *Aur) Install(pkg string) error {
	h, err := a.helperFor()
	if err != nil {
		return err
	}
	return run(a.stdin, a.stdout, a.stderr, h, "-S", "--noconfirm", "--needed", pkg)
}

// Remove removes a package.
func (a *Aur) Remove(pkg string) error {
	h, err := a.helperFor()
	if err != nil {
		return err
	}
	return run(a.stdin, a.stdout, a.stderr, h, "-R", "--noconfirm", pkg)
}

// Upgrade updates a single package from the AUR.
func (a *Aur) Upgrade(pkg string) error {
	h, err := a.helperFor()
	if err != nil {
		return err
	}
	return run(a.stdin, a.stdout, a.stderr, h, "-S", "--noconfirm", "--needed", pkg)
}

// UpgradeAll syncs and updates everything (AUR + repos).
func (a *Aur) UpgradeAll() error {
	h, err := a.helperFor()
	if err != nil {
		return err
	}
	return run(a.stdin, a.stdout, a.stderr, h, "-Syu", "--noconfirm")
}

// List returns all explicitly installed packages.
func (a *Aur) List() ([]string, error) {
	h, err := a.helperFor()
	if err != nil {
		return nil, err
	}
	out, err := query(h, "-Qq")
	if err != nil {
		return nil, err
	}
	return lines(out), nil
}

// Search returns package names matching a term from repos + AUR.
func (a *Aur) Search(term string) ([]string, error) {
	h, err := a.helperFor()
	if err != nil {
		return nil, err
	}
	out, err := query(h, "-Ss", term)
	if err != nil {
		return nil, err
	}
	// Same output shape as `pacman -Ss`: `<repo>/<pkgname> <version> ...`
	// followed by an indented description line.
	var names []string
	for _, l := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(l) == "" || l[0] == ' ' || l[0] == '\t' {
			continue
		}
		name := l
		if i := strings.Index(name, " "); i > 0 {
			name = name[:i]
		}
		if i := strings.Index(name, "/"); i >= 0 {
			name = name[i+1:]
		}
		name = strings.TrimSpace(name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
