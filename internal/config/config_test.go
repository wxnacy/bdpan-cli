package config

import (
    "os"
    "path/filepath"
    "strings"
    "sync"
    "testing"
)

// test helper to reset package state between tests
func resetConfigState() {
    config = nil
    initOnce = sync.Once{}
}

func TestInit_DefaultLoads(t *testing.T) {
    resetConfigState()
    if err := Init(""); err != nil {
        t.Fatalf("Init returned error: %v", err)
    }
    cfg := Get()
    if cfg == nil {
        t.Fatalf("config is nil after Init")
    }
    if cfg.DataDir == "" {
        t.Fatalf("expected DataDir to be set, got empty")
    }
    if strings.Contains(cfg.DataDir, "~") {
        t.Fatalf("expected DataDir to be homedir-expanded, got: %s", cfg.DataDir)
    }
}

func TestInit_MergeFileOverrides(t *testing.T) {
    resetConfigState()
    dir := t.TempDir()
    // use a tilde-based path to verify homedir expansion
    yaml := "data_dir: \"~/.bdtestdir\"\n"
    p := filepath.Join(dir, "conf.yaml")
    if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
        t.Fatalf("write file: %v", err)
    }
    if err := Init(p); err != nil {
        t.Fatalf("Init returned error: %v", err)
    }
    home, _ := os.UserHomeDir()
    expectedPrefix := filepath.Join(home, ".bdtestdir")
    if !strings.HasPrefix(Get().DataDir, expectedPrefix) {
        t.Fatalf("expected DataDir to start with %s, got %s", expectedPrefix, Get().DataDir)
    }
}

func TestInit_OnlyOnce(t *testing.T) {
    resetConfigState()
    dir := t.TempDir()
    p1 := filepath.Join(dir, "a.yaml")
    p2 := filepath.Join(dir, "b.yaml")
    os.WriteFile(p1, []byte("data_dir: \"~/.onlyonce-a\"\n"), 0o644)
    os.WriteFile(p2, []byte("data_dir: \"~/.onlyonce-b\"\n"), 0o644)

    if err := Init(p1); err != nil {
        t.Fatalf("Init p1 error: %v", err)
    }
    first := Get().DataDir

    // second init should be a no-op
    if err := Init(p2); err != nil {
        t.Fatalf("Init p2 error: %v", err)
    }
    second := Get().DataDir

    if first != second {
        t.Fatalf("expected Init to run once; got different DataDir: %s vs %s", first, second)
    }
}

func TestInit_NonYAMLErr(t *testing.T) {
    resetConfigState()
    dir := t.TempDir()
    bad := filepath.Join(dir, "bad.txt")
    os.WriteFile(bad, []byte("not yaml"), 0o644)
    if err := Init(bad); err == nil {
        t.Fatalf("expected error for non-yaml file, got nil")
    }
}

func TestInit_MissingFile_OK(t *testing.T) {
    resetConfigState()
    // nonexistent file should not error; defaults should load
    missing := filepath.Join(t.TempDir(), "missing.yaml")
    if err := Init(missing); err != nil {
        t.Fatalf("expected no error for missing file, got: %v", err)
    }
    if Get().DataDir == "" {
        t.Fatalf("expected DataDir to be set from defaults")
    }
}
