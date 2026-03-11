package app

import "testing"

func TestNormalizeGameKeyStripsRegionTokens(t *testing.T) {
	t.Parallel()

	got := normalizeGameKey("Phantasy Star II (USA, Europe) (Rev A)")
	want := "phantasy star ii (rev a)"
	if got != want {
		t.Fatalf("normalizeGameKey mismatch: got %q want %q", got, want)
	}
}

func TestNormalizeGameKeyPreservesNonRegionGroup(t *testing.T) {
	t.Parallel()

	got := normalizeGameKey("Phantasy Star III (En) (Beta)")
	want := "phantasy star iii (en) (beta)"
	if got != want {
		t.Fatalf("normalizeGameKey mismatch: got %q want %q", got, want)
	}
}

func TestEnsureNoPositionalArgs(t *testing.T) {
	t.Parallel()

	if err := ensureNoPositionalArgs("sync", nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := ensureNoPositionalArgs("sync", []string{"extra"}); err == nil {
		t.Fatal("expected error for positional args")
	}
}
