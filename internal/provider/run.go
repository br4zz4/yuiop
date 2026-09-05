package provider

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// run executes a command with streams attached so interactive prompts work.
func run(stdin io.Reader, stdout, stderr io.Writer, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// query executes a command and returns its combined output.
func query(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	return cmd.CombinedOutput()
}

// sudo prefixes the command with sudo unless already root.
func sudo(command string, args ...string) []string {
	if os.Geteuid() == 0 {
		return append([]string{command}, args...)
	}
	return append([]string{"sudo", command}, args...)
}

// lines splits combined output into trimmed, non-empty lines.
func lines(out []byte) []string {
	var res []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			res = append(res, l)
		}
	}
	return res
}
