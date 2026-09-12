package provider

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// newAurForTest returns an Aur provider whose helper resolution is pinned to
// a fake binary on PATH.
func newAurForTest(t *testing.T, helper string) *Aur {
	t.Helper()
	dir := t.TempDir()
	fake := filepath.Join(dir, helper)
	// A no-op script that exits 0 for every invocation.
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake helper: %v", err)
	}
	t.Setenv("PATH", dir)
	return &Aur{stdin: bytes.NewReader(nil), stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}}
}

func TestAurName(t *testing.T) {
	// arrange
	a := &Aur{}
	// act + assert
	if a.Name() != "aur" {
		t.Fatalf("Name() = %q; want aur", a.Name())
	}
}

func TestAurHelperForPicksYayFirst(t *testing.T) {
	// arrange — both yay and paru present
	dir := t.TempDir()
	for _, h := range []string{"yay", "paru"} {
		p := filepath.Join(dir, h)
		if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write %s: %v", h, err)
		}
	}
	t.Setenv("PATH", dir)
	a := &Aur{}
	// act
	helper, err := a.helperFor()
	// assert
	if err != nil {
		t.Fatalf("helperFor: %v", err)
	}
	if helper != "yay" {
		t.Fatalf("helper = %q; want yay", helper)
	}
}

func TestAurHelperForFallsBackToParu(t *testing.T) {
	// arrange — only paru present
	dir := t.TempDir()
	p := filepath.Join(dir, "paru")
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write paru: %v", err)
	}
	t.Setenv("PATH", dir)
	a := &Aur{}
	// act
	helper, err := a.helperFor()
	// assert
	if err != nil {
		t.Fatalf("helperFor: %v", err)
	}
	if helper != "paru" {
		t.Fatalf("helper = %q; want paru", helper)
	}
}

func TestAurHelperForErrorsWhenNoHelper(t *testing.T) {
	// arrange — empty PATH (no yay/paru)
	t.Setenv("PATH", t.TempDir())
	a := &Aur{}
	// act
	_, err := a.helperFor()
	// assert — error mentions installing yay
	if err == nil {
		t.Fatal("helperFor should error when no helper")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("yay")) {
		t.Fatalf("error should mention yay, got: %v", err)
	}
}

func TestAurInstalledFalseWhenHelperErrors(t *testing.T) {
	// arrange — fake helper exits 0 for -Q (meaning package IS installed)
	a := newAurForTest(t, "yay")
	// act
	installed, err := a.Installed("some-pkg")
	// assert — fake exits 0 → installed
	if err != nil {
		t.Fatalf("Installed: %v", err)
	}
	if !installed {
		t.Fatal("Installed = false; want true (fake helper exits 0)")
	}
}

func TestAurInstalledFalseOnNonZeroExit(t *testing.T) {
	// arrange — fake helper exits 1 for -Q (package not installed)
	dir := t.TempDir()
	fake := filepath.Join(dir, "yay")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write fake: %v", err)
	}
	t.Setenv("PATH", dir)
	a := &Aur{stdin: bytes.NewReader(nil), stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}}
	// act
	installed, err := a.Installed("ghost")
	// assert — non-zero → not installed, no error
	if err != nil {
		t.Fatalf("Installed: %v", err)
	}
	if installed {
		t.Fatal("Installed = true; want false (fake exits 1)")
	}
}

func TestAurInstallUsesHelper(t *testing.T) {
	// arrange
	a := newAurForTest(t, "yay")
	// act — should not error with fake helper on PATH
	if err := a.Install("opencode-bin"); err != nil {
		t.Fatalf("Install: %v", err)
	}
}

func TestAurRemoveUsesHelper(t *testing.T) {
	// arrange
	a := newAurForTest(t, "paru")
	// act
	if err := a.Remove("opencode-bin"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

func TestAurUpgradeUsesHelper(t *testing.T) {
	// arrange
	a := newAurForTest(t, "yay")
	// act
	if err := a.Upgrade("opencode-bin"); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
}

func TestAurUpgradeAllUsesHelper(t *testing.T) {
	// arrange
	a := newAurForTest(t, "yay")
	// act
	if err := a.UpgradeAll(); err != nil {
		t.Fatalf("UpgradeAll: %v", err)
	}
}

func TestAurListReturnsLines(t *testing.T) {
	// arrange — fake helper echoes two packages
	dir := t.TempDir()
	fake := filepath.Join(dir, "yay")
	script := "#!/bin/sh\nif [ \"$1\" = \"-Qq\" ]; then printf 'opencode-bin\\nclaude-code\\n'; else exit 0; fi\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake: %v", err)
	}
	t.Setenv("PATH", dir)
	a := &Aur{stdin: bytes.NewReader(nil), stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}}
	// act
	pkgs, err := a.List()
	// assert
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(pkgs) != 2 || pkgs[0] != "opencode-bin" || pkgs[1] != "claude-code" {
		t.Fatalf("List = %v; want [opencode-bin claude-code]", pkgs)
	}
}

func TestAurSearchParsesNames(t *testing.T) {
	// arrange — fake helper prints pacman-style output with repo/name and descriptions
	dir := t.TempDir()
	fake := filepath.Join(dir, "yay")
	script := `#!/bin/sh
if [ "$1" = "-Ss" ]; then
  printf 'aur/opencode-bin 1.18.30-1\n    Terminal coding agent\ncore/claude-code 2.1.269-1\n    Claude Code CLI\n'
fi
exit 0
`
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake: %v", err)
	}
	t.Setenv("PATH", dir)
	a := &Aur{stdin: bytes.NewReader(nil), stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}}
	// act
	names, err := a.Search("code")
	// assert — only package lines, names without repo prefix
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	want := []string{"opencode-bin", "claude-code"}
	if len(names) != len(want) {
		t.Fatalf("Search = %v; want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("Search[%d] = %q; want %q", i, names[i], want[i])
		}
	}
}
