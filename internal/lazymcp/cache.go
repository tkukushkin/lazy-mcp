package lazymcp

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func CacheDir() string {
	if env := os.Getenv("LAZY_MCP_CACHE_DIR"); env != "" {
		return env
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "lazy-mcp")
}

func CacheKey(command []string) string {
	h := sha256.New()
	for _, part := range command {
		fmt.Fprintf(h, "%d\x00%s", len(part), part)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

type Cache struct {
	path string
}

func NewCache(command []string) *Cache {
	return &Cache{
		path: filepath.Join(CacheDir(), CacheKey(command)+".json"),
	}
}

func (c *Cache) Path() string {
	return c.path
}

func (c *Cache) Load() map[string]json.RawMessage {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

func (c *Cache) Save(data map[string]json.RawMessage) error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return err
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, bytes, 0644)
}
