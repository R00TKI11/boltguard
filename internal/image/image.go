package image

import (
	"context"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// Image wraps container image data we care about
type Image struct {
	Reference string
	Config    *v1.ConfigFile
	Manifest  *v1.Manifest
	Layers    []v1.Layer

	// cached stuff for perf
	manifest *v1.Manifest
	config   *v1.ConfigFile
}

// Load attempts to load an image from local daemon or tarball
// offline=true means we won't try to pull from registry
func Load(ref string, offline bool) (*Image, error) {
	// try daemon first (most common case)
	img, err := loadFromDaemon(ref)
	if err == nil {
		return img, nil
	}

	// maybe it's a tarball path?
	img, err = loadFromTarball(ref)
	if err == nil {
		return img, nil
	}

	if offline {
		return nil, fmt.Errorf("image not found locally and offline mode enabled: %s", ref)
	}

	// TODO: support registry pulls in v0.2
	return nil, fmt.Errorf("remote pulls not yet supported: %s", ref)
}

func loadFromDaemon(ref string) (*Image, error) {
	nameRef, err := name.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	// Use unbuffered opener for better compatibility with Docker Desktop on Windows
	img, err := daemon.Image(nameRef, daemon.WithUnbufferedOpener())
	if err != nil {
		return nil, fmt.Errorf("failed to load from daemon (hint: use 'docker save %s -o image.tar' and scan the tarball): %w", ref, err)
	}

	return buildImage(ref, img)
}

func loadFromTarball(path string) (*Image, error) {
	img, err := tarball.ImageFromPath(path, nil)
	if err != nil {
		return nil, err
	}

	return buildImage(path, img)
}

func buildImage(ref string, img v1.Image) (*Image, error) {
	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to read layers: %w", err)
	}

	return &Image{
		Reference: ref,
		Config:    cfg,
		Manifest:  manifest,
		Layers:    layers,
		manifest:  manifest,
		config:    cfg,
	}, nil
}

// GetFileFromLayers attempts to extract a specific file from the image layers
// This is useful for inspecting /etc/passwd, package manifests, etc.
// Returns io.ReadCloser if found, nil otherwise
func (i *Image) GetFileFromLayers(path string) (io.ReadCloser, error) {
	// TODO: implement layer file extraction
	// For now we'll just support facts that don't need this
	return nil, fmt.Errorf("layer file extraction not yet implemented")
}

// Size returns the total image size in bytes
func (i *Image) Size() (int64, error) {
	var total int64
	for _, layer := range i.Layers {
		size, err := layer.Size()
		if err != nil {
			return 0, err
		}
		total += size
	}
	return total, nil
}

// DiffIDs returns the layer diff IDs
func (i *Image) DiffIDs() ([]v1.Hash, error) {
	if i.config == nil {
		return nil, fmt.Errorf("no config available")
	}
	return i.config.RootFS.DiffIDs, nil
}

// Inspect returns image details for debugging
func (i *Image) Inspect(ctx context.Context) map[string]interface{} {
	info := map[string]interface{}{
		"reference": i.Reference,
	}

	if i.config != nil {
		info["created"] = i.config.Created
		info["architecture"] = i.config.Architecture
		info["os"] = i.config.OS

		if i.config.Config.User != "" {
			info["user"] = i.config.Config.User
		}

		if len(i.config.Config.Env) > 0 {
			info["env_count"] = len(i.config.Config.Env)
		}

		if len(i.config.Config.ExposedPorts) > 0 {
			info["exposed_ports"] = len(i.config.Config.ExposedPorts)
		}
	}

	size, err := i.Size()
	if err == nil {
		info["size_bytes"] = size
	}

	return info
}
