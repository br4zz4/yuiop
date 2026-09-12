package self

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTarget(t *testing.T) {
	// arrange + act
	target, err := DetectTarget()
	// assert
	if err != nil {
		t.Fatalf("DetectTarget: %v", err)
	}
	// Only darwin/linux supported; arch must be amd64 or arm64.
	if target != "darwin_amd64" && target != "darwin_arm64" && target != "linux_amd64" && target != "linux_arm64" {
		t.Fatalf("DetectTarget = %q; want a supported target", target)
	}
}

func TestFetchLatestVersion(t *testing.T) {
	// arrange — fake GitHub API
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"tag_name":"v0.2.0","name":"v0.2.0"}`)
	}))
	defer srv.Close()
	// patch the package-level api URL for the test
	origAPI := api
	api = srv.URL + "/repos/" + repo
	defer func() { api = origAPI }()

	client := srv.Client()
	// act
	version, err := FetchLatestVersion(client)
	// assert
	if err != nil {
		t.Fatalf("FetchLatestVersion: %v", err)
	}
	if version != "v0.2.0" {
		t.Fatalf("version = %q; want v0.2.0", version)
	}
}

func TestDownloadToWritesBinary(t *testing.T) {
	// arrange — fake release server serving the asset
	payload := []byte("#!/bin/sh\necho new-yuiop\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/br4zz4/yuiop/releases/download/v0.2.0/yuiop_0.2.0_darwin_arm64" {
			w.Write(payload)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	// point the download base at the fake server
	origBase := downloadBase
	downloadBase = srv.URL
	defer func() { downloadBase = origBase }()

	client := srv.Client()
	// act
	path, err := DownloadTo("v0.2.0", "darwin_arm64", client)
	// assert
	if err != nil {
		t.Fatalf("DownloadTo: %v", err)
	}
	defer os.Remove(path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp: %v", err)
	}
	if string(data) != string(payload) {
		t.Fatalf("content = %q; want %q", data, payload)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatal("downloaded binary is not executable")
	}
}

func TestDownloadToErrorsOn404(t *testing.T) {
	// arrange — server returns 404 for everything
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()
	client := srv.Client()
	// act
	_, err := DownloadTo("v9.9.9", "darwin_arm64", client)
	// assert
	if err == nil {
		t.Fatal("DownloadTo should error on 404")
	}
}

func TestReplaceAtSwapsBinary(t *testing.T) {
	// arrange — a fake "installed" binary and a fake "new" binary
	dir := t.TempDir()
	exe := filepath.Join(dir, "yuiop")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}
	newBin := filepath.Join(dir, "new-yuiop")
	if err := os.WriteFile(newBin, []byte("new"), 0o755); err != nil {
		t.Fatalf("write new: %v", err)
	}
	// act
	replaced, err := ReplaceAt(newBin, exe)
	// assert — exe now contains new content, newBin is gone
	if err != nil {
		t.Fatalf("ReplaceAt: %v", err)
	}
	if replaced != exe {
		t.Fatalf("replaced = %q; want %q", replaced, exe)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read exe: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("exe content = %q; want new", data)
	}
	if _, err := os.Stat(newBin); !os.IsNotExist(err) {
		t.Fatalf("newBin should be gone after rename, stat err = %v", err)
	}
}

func TestReplaceAtErrorsWhenSourceMissing(t *testing.T) {
	// arrange
	dir := t.TempDir()
	exe := filepath.Join(dir, "yuiop")
	// act
	_, err := ReplaceAt(filepath.Join(dir, "does-not-exist"), exe)
	// assert
	if err == nil {
		t.Fatal("ReplaceAt should error for a missing new binary")
	}
}
