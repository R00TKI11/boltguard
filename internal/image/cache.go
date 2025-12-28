package image

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// CachedResult stores scan results keyed by image digest
type CachedResult struct {
	Digest    string                 `json:"digest"`
	ImageRef  string                 `json:"image_ref"`
	Config    *v1.ConfigFile         `json:"config"`
	Manifest  *v1.Manifest           `json:"manifest"`
	Size      int64                  `json:"size"`
	LayerInfo []CachedLayer          `json:"layers"`
	CachedAt  time.Time              `json:"cached_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type CachedLayer struct {
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

// Cache manages persistent image scan cache
type Cache struct {
	dir     string
	enabled bool
}

// NewCache creates a cache instance
// if dir is empty, uses default location
func NewCache(dir string, enabled bool) (*Cache, error) {
	if !enabled {
		return &Cache{enabled: false}, nil
	}

	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		dir = filepath.Join(home, ".cache", "boltguard")
	}

	// create cache dir if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}

	return &Cache{
		dir:     dir,
		enabled: true,
	}, nil
}

// Get retrieves a cached result by digest
func (c *Cache) Get(digest string) (*CachedResult, bool) {
	if !c.enabled {
		return nil, false
	}

	path := c.cachePath(digest)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var result CachedResult
	if err := json.Unmarshal(data, &result); err != nil {
		// corrupted cache, ignore it
		return nil, false
	}

	return &result, true
}

// Put stores a result in the cache
func (c *Cache) Put(digest string, result *CachedResult) error {
	if !c.enabled {
		return nil
	}

	result.CachedAt = time.Now()
	result.Digest = digest

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	path := c.cachePath(digest)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes all cached entries
func (c *Cache) Clear() error {
	if !c.enabled {
		return nil
	}

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(c.dir, entry.Name())
			if err := os.Remove(path); err != nil {
				// log but don't fail
				fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", path, err)
			}
		}
	}

	return nil
}

// Prune removes cache entries older than the given duration
func (c *Cache) Prune(maxAge time.Duration) error {
	if !c.enabled {
		return nil
	}

	cutoff := time.Now().Add(-maxAge)
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache dir: %w", err)
	}

	pruned := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(c.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var result CachedResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		if result.CachedAt.Before(cutoff) {
			if err := os.Remove(path); err == nil {
				pruned++
			}
		}
	}

	if pruned > 0 {
		fmt.Fprintf(os.Stderr, "pruned %d old cache entries\n", pruned)
	}

	return nil
}

// Stats returns cache statistics
func (c *Cache) Stats() (int, int64, error) {
	if !c.enabled {
		return 0, 0, nil
	}

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read cache dir: %w", err)
	}

	var count int
	var totalSize int64

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			count++
			totalSize += info.Size()
		}
	}

	return count, totalSize, nil
}

func (c *Cache) cachePath(digest string) string {
	// sanitize digest for filesystem
	safe := fmt.Sprintf("%x", sha256.Sum256([]byte(digest)))
	return filepath.Join(c.dir, safe+".json")
}

// ImageToCache converts an Image to a CachedResult
func ImageToCache(img *Image) (*CachedResult, error) {
	var layers []CachedLayer
	for _, layer := range img.Layers {
		digest, _ := layer.Digest()
		size, _ := layer.Size()
		layers = append(layers, CachedLayer{
			Digest: digest.String(),
			Size:   size,
		})
	}

	size, err := img.Size()
	if err != nil {
		return nil, err
	}

	return &CachedResult{
		ImageRef:  img.Reference,
		Config:    img.Config,
		Manifest:  img.Manifest,
		Size:      size,
		LayerInfo: layers,
	}, nil
}
