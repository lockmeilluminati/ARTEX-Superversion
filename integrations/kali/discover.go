// Package kali provides read-only discovery of a small, fixed set of tools.
// It never invokes a shell, accepts user-supplied commands, installs packages,
// or runs scans. Each binary is called only with its fixed version flag.
package kali

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

type Tool struct {
	Name      string    `json:"name"`
	Path      string    `json:"path,omitempty"`
	Installed bool      `json:"installed"`
	Version   string    `json:"version,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
	Diagnostic string   `json:"diagnostic,omitempty"`
}

type toolSpec struct {
	name string
	args []string
}

var selectedTools = []toolSpec{
	{name: "nmap", args: []string{"--version"}},
	{name: "nuclei", args: []string{"-version"}},
	{name: "ffuf", args: []string{"-V"}},
	{name: "sqlmap", args: []string{"--version"}},
	{name: "nikto", args: []string{"-Version"}},
}

const maxVersionOutput = 4096

// Discover checks only the fixed allowlist above. Missing tools are normal and
// are reported without making ARTEX startup fail.
func Discover(ctx context.Context) []Tool {
	out := make([]Tool, 0, len(selectedTools))
	for _, spec := range selectedTools {
		item := Tool{Name: spec.name, CheckedAt: time.Now().UTC()}
		path, err := exec.LookPath(spec.name)
		if err != nil {
			item.Diagnostic = "not found on PATH"
			out = append(out, item)
			continue
		}
		item.Path, item.Installed = path, true
		callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		cmd := exec.CommandContext(callCtx, path, spec.args...)
		var output limitedBuffer
		cmd.Stdout, cmd.Stderr = &output, &output
		err = cmd.Run()
		cancel()
		item.Version = cleanVersion(output.String())
		if err != nil && item.Version == "" {
			item.Diagnostic = "version check failed"
		}
		out = append(out, item)
	}
	return out
}

type limitedBuffer struct {
	buf bytes.Buffer
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := maxVersionOutput - b.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) String() string { return b.buf.String() }

func cleanVersion(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > maxVersionOutput {
		value = value[:maxVersionOutput]
	}
	return value
}
