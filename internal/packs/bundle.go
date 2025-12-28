package packs

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Bundle represents a policy/advisory pack for offline updates
type Bundle struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Policies    []Policy  `json:"policies"`
	Advisories  []Advisory `json:"advisories,omitempty"`
}

type Policy struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type Advisory struct {
	ID          string   `json:"id"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Affected    []string `json:"affected"` // base images affected
	Remediation string   `json:"remediation"`
}

// BundleManager handles importing and managing bundles
type BundleManager struct {
	dir string
}

// NewBundleManager creates a bundle manager
func NewBundleManager(dir string) (*BundleManager, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		dir = filepath.Join(home, ".config", "boltguard", "packs")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create packs dir: %w", err)
	}

	return &BundleManager{dir: dir}, nil
}

// Import loads a bundle tarball and extracts it
//nolint:errcheck // defer close calls - standard pattern
func (m *BundleManager) Import(path string) (*Bundle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress bundle: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var bundle *Bundle
	policies := make(map[string]string)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// read manifest
		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}

			var b Bundle
			if err := json.Unmarshal(data, &b); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %w", err)
			}
			bundle = &b
			continue
		}

		// read policy files
		if filepath.Ext(header.Name) == ".yaml" || filepath.Ext(header.Name) == ".yml" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read policy %s: %w", header.Name, err)
			}
			policies[header.Name] = string(data)
		}
	}

	if bundle == nil {
		return nil, fmt.Errorf("bundle missing manifest.json")
	}

	// associate policy content
	for i := range bundle.Policies {
		if content, ok := policies[bundle.Policies[i].Filename]; ok {
			bundle.Policies[i].Content = content
		}
	}

	// save bundle to local store
	if err := m.save(bundle); err != nil {
		return nil, fmt.Errorf("failed to save bundle: %w", err)
	}

	return bundle, nil
}

// List returns all installed bundles
func (m *BundleManager) List() ([]Bundle, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read packs dir: %w", err)
	}

	var bundles []Bundle
	for _, entry := range entries {
		if entry.IsDir() {
			manifestPath := filepath.Join(m.dir, entry.Name(), "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}

			var b Bundle
			if err := json.Unmarshal(data, &b); err != nil {
				continue
			}
			bundles = append(bundles, b)
		}
	}

	return bundles, nil
}

// Get retrieves a specific bundle by name
func (m *BundleManager) Get(name string) (*Bundle, error) {
	manifestPath := filepath.Join(m.dir, name, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("bundle not found: %s", name)
	}

	var b Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", err)
	}

	// load policy content
	for i := range b.Policies {
		policyPath := filepath.Join(m.dir, name, b.Policies[i].Filename)
		content, err := os.ReadFile(policyPath)
		if err != nil {
			continue
		}
		b.Policies[i].Content = string(content)
	}

	return &b, nil
}

// GetPolicy retrieves a specific policy from a bundle
func (m *BundleManager) GetPolicy(bundleName, policyName string) (string, error) {
	bundle, err := m.Get(bundleName)
	if err != nil {
		return "", err
	}

	for _, policy := range bundle.Policies {
		if policy.Name == policyName {
			return policy.Content, nil
		}
	}

	return "", fmt.Errorf("policy not found: %s", policyName)
}

// Remove deletes a bundle
func (m *BundleManager) Remove(name string) error {
	bundleDir := filepath.Join(m.dir, name)
	return os.RemoveAll(bundleDir)
}

func (m *BundleManager) save(bundle *Bundle) error {
	bundleDir := filepath.Join(m.dir, bundle.Name)
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return err
	}

	// save manifest
	manifestPath := filepath.Join(bundleDir, "manifest.json")
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return err
	}

	// save policies
	for _, policy := range bundle.Policies {
		policyPath := filepath.Join(bundleDir, policy.Filename)
		if err := os.WriteFile(policyPath, []byte(policy.Content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// Export creates a bundle tarball from a directory
//nolint:errcheck // defer close calls - standard pattern
func Export(sourceDir, outputPath, name, version, description string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// collect policies
	var policies []Policy
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(sourceDir, entry.Name()))
		if err != nil {
			return err
		}

		policies = append(policies, Policy{
			Name:     entry.Name(),
			Filename: entry.Name(),
			Content:  string(content),
		})

		// add to tar
		if err := tw.WriteHeader(&tar.Header{
			Name:    entry.Name(),
			Size:    int64(len(content)),
			Mode:    0644,
			ModTime: time.Now(),
		}); err != nil {
			return err
		}
		if _, err := tw.Write(content); err != nil {
			return err
		}
	}

	// create manifest
	bundle := Bundle{
		Name:        name,
		Version:     version,
		Description: description,
		CreatedAt:   time.Now(),
		Policies:    policies,
	}

	manifestData, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}

	// add manifest to tar
	if err := tw.WriteHeader(&tar.Header{
		Name:    "manifest.json",
		Size:    int64(len(manifestData)),
		Mode:    0644,
		ModTime: time.Now(),
	}); err != nil {
		return err
	}
	if _, err := tw.Write(manifestData); err != nil {
		return err
	}

	return nil
}
