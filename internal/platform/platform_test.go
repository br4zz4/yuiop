package platform

import "testing"

func TestValidate(t *testing.T) {
	for _, ok := range []string{"brew", "apt", "pacman"} {
		if _, err := Validate(ok); err != nil {
			t.Fatalf("Validate(%q) = %v; want nil", ok, err)
		}
	}
	if _, err := Validate("windows"); err == nil {
		t.Fatal("Validate(windows) should fail")
	}
}

func TestResolvePrecedence(t *testing.T) {
	if p, _ := Resolve("brew", "apt", "pacman"); p != "brew" {
		t.Fatalf("cli should win, got %q", p)
	}
	if p, _ := Resolve("", "apt", "pacman"); p != "apt" {
		t.Fatalf("env should beat config, got %q", p)
	}
	if p, _ := Resolve("", "", "pacman"); p != "pacman" {
		t.Fatalf("config should be used, got %q", p)
	}
}