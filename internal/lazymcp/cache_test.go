package lazymcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCacheKey(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		k1 := CacheKey([]string{"a", "b"})
		k2 := CacheKey([]string{"a", "b"})
		if k1 != k2 {
			t.Fatal("same input must produce same key")
		}
	})

	t.Run("different commands", func(t *testing.T) {
		k1 := CacheKey([]string{"a", "b"})
		k2 := CacheKey([]string{"a", "c"})
		if k1 == k2 {
			t.Fatal("different input must produce different keys")
		}
	})

	t.Run("order matters", func(t *testing.T) {
		k1 := CacheKey([]string{"a", "b"})
		k2 := CacheKey([]string{"b", "a"})
		if k1 == k2 {
			t.Fatal("order must matter")
		}
	})

	t.Run("separator collision", func(t *testing.T) {
		k1 := CacheKey([]string{"a\x00b"})
		k2 := CacheKey([]string{"a", "b"})
		if k1 == k2 {
			t.Fatal("must not collide with null separator")
		}
	})
}

func TestCacheDir(t *testing.T) {
	t.Run("env override", func(t *testing.T) {
		t.Setenv("LAZY_MCP_CACHE_DIR", "/custom/path")
		dir := CacheDir()
		if dir != "/custom/path" {
			t.Fatalf("got %q, want /custom/path", dir)
		}
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("LAZY_MCP_CACHE_DIR", "")
		dir := CacheDir()
		if dir == "" {
			t.Fatal("default cache dir must not be empty")
		}
	})
}

func TestCache(t *testing.T) {
	t.Run("load no file", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("LAZY_MCP_CACHE_DIR", dir)
		c := NewCache([]string{"test"})
		data := c.Load()
		if data != nil {
			t.Fatal("load on missing file must return nil")
		}
	})

	t.Run("save and load", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("LAZY_MCP_CACHE_DIR", dir)
		c := NewCache([]string{"test"})
		input := map[string]json.RawMessage{
			"initialize": json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
		}
		if err := c.Save(input); err != nil {
			t.Fatal(err)
		}
		loaded := c.Load()
		if loaded == nil {
			t.Fatal("loaded must not be nil")
		}
		if string(loaded["initialize"]) != `{"protocolVersion":"2024-11-05"}` {
			t.Fatalf("unexpected: %s", loaded["initialize"])
		}
	})

	t.Run("corrupted file", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("LAZY_MCP_CACHE_DIR", dir)
		c := NewCache([]string{"test"})
		path := filepath.Join(dir, CacheKey([]string{"test"})+".json")
		os.WriteFile(path, []byte("not valid json{{{"), 0644)
		data := c.Load()
		if data != nil {
			t.Fatal("corrupted file must return nil")
		}
	})

	t.Run("creates parent dirs", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "deep", "nested")
		t.Setenv("LAZY_MCP_CACHE_DIR", dir)
		c := NewCache([]string{"test"})
		input := map[string]json.RawMessage{"test": json.RawMessage(`true`)}
		if err := c.Save(input); err != nil {
			t.Fatal(err)
		}
		loaded := c.Load()
		if loaded == nil {
			t.Fatal("must load after save")
		}
	})
}
