package app

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestResolveConfigPathsLayering(t *testing.T) {
	tmp := t.TempDir()

	configHome := filepath.Join(tmp, "xdg")
	projectDir := filepath.Join(tmp, "project")
	override := filepath.Join(tmp, "override.yaml")

	if err := os.MkdirAll(filepath.Join(configHome, "retro-collection-tool"), 0o755); err != nil {
		t.Fatalf("mkdir xdg: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "config"), 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	xdgCfg := filepath.Join(configHome, "retro-collection-tool", "config.yaml")
	projectCfg := filepath.Join(projectDir, "config", "retro-collection-tool.yaml")

	if err := os.WriteFile(xdgCfg, []byte("root: /x\n"), 0o644); err != nil {
		t.Fatalf("write xdg cfg: %v", err)
	}
	if err := os.WriteFile(projectCfg, []byte("paths: {}\n"), 0o644); err != nil {
		t.Fatalf("write project cfg: %v", err)
	}
	if err := os.WriteFile(override, []byte("cache_dir: cache\n"), 0o644); err != nil {
		t.Fatalf("write override cfg: %v", err)
	}

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origCwd)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv(configEnvVar, override)

	paths, err := resolveConfigPaths("")
	if err != nil {
		t.Fatalf("resolveConfigPaths: %v", err)
	}

	if len(paths) != 3 {
		t.Fatalf("expected 3 config layers, got %d: %#v", len(paths), paths)
	}
	if paths[0] != xdgCfg || paths[1] != projectCfg || paths[2] != override {
		t.Fatalf("unexpected layering order: %#v", paths)
	}
}
