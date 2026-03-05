package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EnvPair is a key=value pair.
type EnvPair struct {
	Key   string
	Value string
}

// ParseLines parses key=value lines, ignoring blank lines and # comments.
func ParseLines(content string) ([]EnvPair, error) {
	var pairs []EnvPair
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			return nil, fmt.Errorf("line %d: expected KEY=VALUE, got %q", i+1, line)
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		// Strip surrounding quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		pairs = append(pairs, EnvPair{Key: key, Value: value})
	}
	return pairs, nil
}

// FormatLines formats env pairs as KEY=VALUE lines.
func FormatLines(pairs []EnvPair) string {
	var sb strings.Builder
	for _, p := range pairs {
		fmt.Fprintf(&sb, "%s=%s\n", p.Key, p.Value)
	}
	return sb.String()
}

// Open opens $EDITOR with initial content, returns edited pairs.
func Open(initial []EnvPair) ([]EnvPair, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	f, err := os.CreateTemp("", "umgebung-*.env")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())

	f.WriteString("# umgebung - one KEY=VALUE per line, blank lines and # comments ignored\n")
	f.WriteString(FormatLines(initial))
	f.Close()

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor: %w", err)
	}

	content, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}
	return ParseLines(string(content))
}
