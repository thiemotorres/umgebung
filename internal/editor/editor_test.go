package editor_test

import (
	"testing"

	"github.com/thiemotorres/umgebung/internal/editor"
)

func TestParseLines(t *testing.T) {
	input := `
# comment
FOO=bar
BAR=hello world
BAZ="quoted value"
EMPTY=
`
	pairs, err := editor.ParseLines(input)
	if err != nil {
		t.Fatalf("ParseLines: %v", err)
	}
	if len(pairs) != 4 {
		t.Fatalf("want 4 pairs, got %d: %v", len(pairs), pairs)
	}
	cases := []editor.EnvPair{
		{Key: "FOO", Value: "bar"},
		{Key: "BAR", Value: "hello world"},
		{Key: "BAZ", Value: "quoted value"},
		{Key: "EMPTY", Value: ""},
	}
	for i, want := range cases {
		if pairs[i] != want {
			t.Errorf("pair %d: got %v, want %v", i, pairs[i], want)
		}
	}
}

func TestFormatLines(t *testing.T) {
	pairs := []editor.EnvPair{
		{Key: "FOO", Value: "bar"},
		{Key: "BAR", Value: "baz"},
	}
	got := editor.FormatLines(pairs)
	want := "FOO=bar\nBAR=baz\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseLineShellSubstitution(t *testing.T) {
	input := `FOO=$(printf 'hello')`
	pairs, err := editor.ParseLines(input)
	if err != nil {
		t.Fatalf("ParseLines: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Value != "hello" {
		t.Errorf("expected value %q, got %q", "hello", pairs[0].Value)
	}
}

func TestParseLineShellSubstitutionMultiline(t *testing.T) {
	// printf with \n produces a real newline - store it as-is
	input := "FOO=$(printf 'line1\\nline2')"
	pairs, err := editor.ParseLines(input)
	if err != nil {
		t.Fatalf("ParseLines: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Value != "line1\nline2" {
		t.Errorf("expected multiline value, got %q", pairs[0].Value)
	}
}

func TestParseLineLiteralValueUnchanged(t *testing.T) {
	// Plain values without $() must not be altered
	input := `FOO=plainvalue`
	pairs, err := editor.ParseLines(input)
	if err != nil {
		t.Fatalf("ParseLines: %v", err)
	}
	if pairs[0].Value != "plainvalue" {
		t.Errorf("expected %q, got %q", "plainvalue", pairs[0].Value)
	}
}
