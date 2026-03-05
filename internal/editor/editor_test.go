package editor_test

import (
	"testing"

	"github.com/feto/umgebung/internal/editor"
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
