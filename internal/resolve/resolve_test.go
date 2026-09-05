package resolve

import "testing"

func TestResolveKnown(t *testing.T) {
	tab, err := Load(embeddedYAML)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	pkg, ok := tab.Resolve("delta", "brew")
	if !ok || pkg != "git-delta" {
		t.Fatalf("delta/brew = %q, %v; want git-delta, true", pkg, ok)
	}
	if _, ok := tab.Resolve("opencode", "apt"); ok {
		t.Fatal("opencode/apt should be unknown")
	}
}

func TestResolveUnknown(t *testing.T) {
	tab, err := Load(embeddedYAML)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, ok := tab.Resolve("no-such-package", "brew"); ok {
		t.Fatal("unknown canonical should not resolve")
	}
}

var embeddedYAML = []byte(`
packages:
  delta: { brew: git-delta, apt: git-delta, pacman: git-delta }
  opencode: { brew: anomalyco/tap/opencode }
`)