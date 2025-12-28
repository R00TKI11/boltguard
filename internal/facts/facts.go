package facts

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/yourusername/boltguard/internal/image"
)

// Facts represents everything we extract from an image that policies care about
type Facts struct {
	// Basic metadata
	BaseImage    string
	Size         int64
	Created      time.Time
	Architecture string
	OS           string

	// User/permissions
	User         string
	RunsAsRoot   bool
	HasSetuidBit bool // TODO: requires layer scanning

	// Labels
	Labels map[string]string

	// Config details
	Env          []string
	ExposedPorts []string
	Entrypoint   []string
	Cmd          []string
	WorkingDir   string

	// Layer info
	LayerCount int
	Layers     []LayerFact

	// History
	History []v1.History

	// Files (requires deeper inspection, maybe v0.2)
	// PackageManagers []string
	// InstalledPackages []Package
}

type LayerFact struct {
	Digest    string
	Size      int64
	CreatedBy string
}

// Extract pulls out all the facts from an image
func Extract(img *image.Image) (*Facts, error) {
	f := &Facts{
		Labels: make(map[string]string),
	}

	if img.Config == nil {
		return nil, fmt.Errorf("image has no config")
	}

	cfg := img.Config

	// basic metadata
	f.Created = cfg.Created.Time
	f.Architecture = cfg.Architecture
	f.OS = cfg.OS

	// user info
	f.User = cfg.Config.User
	if f.User == "" || f.User == "root" || f.User == "0" {
		f.RunsAsRoot = true
	}

	// labels
	if cfg.Config.Labels != nil {
		f.Labels = cfg.Config.Labels
	}

	// env
	f.Env = cfg.Config.Env

	// ports
	for port := range cfg.Config.ExposedPorts {
		f.ExposedPorts = append(f.ExposedPorts, port)
	}

	// cmd/entrypoint
	f.Entrypoint = cfg.Config.Entrypoint
	f.Cmd = cfg.Config.Cmd
	f.WorkingDir = cfg.Config.WorkingDir

	// layers
	f.LayerCount = len(img.Layers)
	for _, layer := range img.Layers {
		digest, _ := layer.Digest()
		size, _ := layer.Size()

		f.Layers = append(f.Layers, LayerFact{
			Digest: digest.String(),
			Size:   size,
		})
	}

	// history
	if cfg.History != nil {
		f.History = cfg.History
	}

	// size
	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate size: %w", err)
	}
	f.Size = size

	// try to infer base image from history
	f.BaseImage = inferBaseImage(cfg.History)

	return f, nil
}

// inferBaseImage attempts to guess the base image from history
// This is kinda hacky but works for most common cases
func inferBaseImage(history []v1.History) string {
	if len(history) == 0 {
		return "unknown"
	}

	// look for FROM commands in history
	fromRegex := regexp.MustCompile(`(?i)FROM\s+([^\s]+)`)

	for _, h := range history {
		if h.CreatedBy == "" {
			continue
		}

		matches := fromRegex.FindStringSubmatch(h.CreatedBy)
		if len(matches) > 1 {
			return matches[1]
		}

		// sometimes it's in a comment
		if strings.Contains(strings.ToLower(h.CreatedBy), "from") {
			parts := strings.Fields(h.CreatedBy)
			for i, part := range parts {
				if strings.ToLower(part) == "from" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}

	return "unknown"
}

// GetLabel safely retrieves a label value
func (f *Facts) GetLabel(key string) string {
	return f.Labels[key]
}

// HasLabel checks if a label exists
func (f *Facts) HasLabel(key string) bool {
	_, exists := f.Labels[key]
	return exists
}

// GetEnvVar retrieves an environment variable value
func (f *Facts) GetEnvVar(key string) string {
	prefix := key + "="
	for _, env := range f.Env {
		if strings.HasPrefix(env, prefix) {
			return strings.TrimPrefix(env, prefix)
		}
	}
	return ""
}

// HasEnvVar checks if an env var is set
func (f *Facts) HasEnvVar(key string) bool {
	prefix := key + "="
	for _, env := range f.Env {
		if strings.HasPrefix(env, prefix) {
			return true
		}
	}
	return false
}

// SizeMB returns size in megabytes
func (f *Facts) SizeMB() float64 {
	return float64(f.Size) / (1024 * 1024)
}

// SizeGB returns size in gigabytes
func (f *Facts) SizeGB() float64 {
	return float64(f.Size) / (1024 * 1024 * 1024)
}
