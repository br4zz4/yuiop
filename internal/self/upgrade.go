// Package self implements `yuiop self upgrade` — fetching the latest release
// binary and replacing the running executable in place.
package self

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repo = "br4zz4/yuiop"
)

// api is the GitHub API base; overridable in tests.
var api = "https://api.github.com/repos/" + repo

// downloadBase is the GitHub host for release assets; overridable in tests.
var downloadBase = "https://github.com"

// DetectTarget maps the running OS/arch to the release asset suffix
// (e.g. darwin_arm64, linux_amd64).
func DetectTarget() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("unsupported arch %q", arch)
	}
	switch osName {
	case "darwin", "linux":
	default:
		return "", fmt.Errorf("unsupported OS %q", osName)
	}
	return osName + "_" + arch, nil
}

// FetchLatestVersion returns the tag name of the latest release (e.g. v0.2.0).
func FetchLatestVersion(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, api+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "yuiop-self-upgrade")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch latest release: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// {"tag_name":"v0.2.0", ...} — parse by hand, no JSON dep.
	bodyStr := string(body)
	idx := strings.Index(bodyStr, `"tag_name":"`)
	if idx < 0 {
		return "", fmt.Errorf("no tag_name in release response")
	}
	rest := bodyStr[idx+len(`"tag_name":"`):]
	close := strings.Index(rest, `"`)
	if close < 0 {
		return "", fmt.Errorf("malformed tag_name")
	}
	return rest[:close], nil
}

// DownloadTo downloads the release asset for version/target into a temp file
// and returns its path.
func DownloadTo(version, target string, client *http.Client) (string, error) {
	url := fmt.Sprintf("%s/%s/releases/download/%s/yuiop_%s_%s",
		downloadBase, repo, version, strings.TrimPrefix(version, "v"), target)
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("yuiop-upgrade-%d", time.Now().UnixNano()))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "yuiop-self-upgrade")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(tmp)
		return "", err
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return tmp, nil
}

// ReplaceExecutable swaps the running binary with the downloaded one.
// Returns the path that was replaced.
func ReplaceExecutable(newBin string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	// Resolve symlinks (e.g. ~/.local/bin/yuiop → real location) so we replace
	// the actual file, not the link.
	real, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = real
	}
	return ReplaceAt(newBin, exe)
}

// ReplaceAt atomically moves newBin over exePath. Split out for testability.
// Falls back to copy+remove when rename fails — e.g. EXDEV (cross-device link)
// when /tmp is tmpfs and the binary lives on a real disk.
func ReplaceAt(newBin, exePath string) (string, error) {
	if err := os.Rename(newBin, exePath); err == nil {
		return exePath, nil
	}
	if err := copyFile(newBin, exePath); err != nil {
		return "", fmt.Errorf("replace %s: %w", exePath, err)
	}
	if err := os.Chmod(exePath, 0o755); err != nil {
		return "", fmt.Errorf("replace %s: %w", exePath, err)
	}
	os.Remove(newBin)
	return exePath, nil
}

// copyFile streams src onto dst (truncating dst), creating parent dirs.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if parent := filepath.Dir(dst); parent != "" && parent != "." {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return err
		}
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
