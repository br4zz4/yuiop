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

func TestResolveOrSelfAurUsesCanonicalName(t *testing.T) {
	// arrange
	tab, err := Load(embeddedYAML)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// act — opencode-bin isn't in the table, but AUR accepts any name
	pkg, ok := tab.ResolveOrSelf("opencode-bin", "aur")
	// assert
	if !ok || pkg != "opencode-bin" {
		t.Fatalf("ResolveOrSelf(opencode-bin, aur) = %q, %v; want opencode-bin, true", pkg, ok)
	}
}

func TestResolveOrSelfKnownMappingStillWins(t *testing.T) {
	// arrange
	tab, err := Load(embeddedYAML)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// act — delta has a brew mapping; aur not mapped → falls back to canonical
	pkg, ok := tab.ResolveOrSelf("delta", "brew")
	// assert
	if !ok || pkg != "git-delta" {
		t.Fatalf("ResolveOrSelf(delta, brew) = %q, %v; want git-delta, true", pkg, ok)
	}
}

func TestResolveOrSelfUnknownProviderFails(t *testing.T) {
	// arrange
	tab, err := Load(embeddedYAML)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// act — no mapping and not aur → fails
	if _, ok := tab.ResolveOrSelf("no-such-package", "brew"); ok {
		t.Fatal("ResolveOrSelf(no-such-package, brew) should not resolve")
	}
}
