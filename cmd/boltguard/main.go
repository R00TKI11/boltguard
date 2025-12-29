package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/R00TKI11/boltguard/internal/facts"
	"github.com/R00TKI11/boltguard/internal/image"
	"github.com/R00TKI11/boltguard/internal/packs"
	"github.com/R00TKI11/boltguard/internal/policy"
	"github.com/R00TKI11/boltguard/internal/report"
	"github.com/R00TKI11/boltguard/internal/rules"
)

const version = "0.1.0"

func main() {
	// flags
	var (
		policyFile   = flag.String("policy", "", "path to policy file (defaults to built-in)")
		outputFormat = flag.String("format", "text", "output format: text, json, sarif")
		verbose      = flag.Bool("v", false, "verbose output")
		showVersion  = flag.Bool("version", false, "print version and exit")
		offline      = flag.Bool("offline", true, "operate in offline mode (default true)")

		// cache flags
		useCache   = flag.Bool("cache", true, "use cache for faster scans")
		cacheDir   = flag.String("cache-dir", "", "cache directory (default: ~/.cache/boltguard)")
		clearCache = flag.Bool("cache-clear", false, "clear cache and exit")
		cacheStats = flag.Bool("cache-stats", false, "show cache stats and exit")

		// bundle flags
		bundleImport = flag.String("bundle-import", "", "import policy bundle (.tar.gz)")
		bundleList   = flag.Bool("bundle-list", false, "list installed bundles")
		bundleExport = flag.String("bundle-export", "", "export policies as bundle")
	)

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: boltguard [options] <image>\n\n")
		fmt.Fprintf(os.Stderr, "BoltGuard - Fast, offline container policy checks\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  boltguard nginx:latest\n")
		fmt.Fprintf(os.Stderr, "  boltguard -policy custom.yaml alpine:3.18\n")
		fmt.Fprintf(os.Stderr, "  boltguard -format json redis:7 > report.json\n")
		fmt.Fprintf(os.Stderr, "  boltguard -bundle-import policies.tar.gz\n")
		fmt.Fprintf(os.Stderr, "  boltguard -cache-clear\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("boltguard v%s\n", version)
		os.Exit(0)
	}

	// handle cache operations
	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if *clearCache {
		if err := handleCacheClear(*cacheDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if *cacheStats {
		if err := handleCacheStats(*cacheDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// handle bundle operations
	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if *bundleImport != "" {
		if err := handleBundleImport(*bundleImport); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if *bundleList {
		if err := handleBundleList(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if *bundleExport != "" {
		if err := handleBundleExport(*bundleExport); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// normal scan mode
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	imageName := flag.Arg(0)

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if err := run(imageName, *policyFile, *outputFormat, *verbose, *offline, *useCache, *cacheDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(imageName, policyPath, format string, verbose, offline, useCache bool, cacheDir string) error {
	// init cache if enabled
	cache, err := image.NewCache(cacheDir, useCache)
	if err != nil {
		return fmt.Errorf("failed to init cache: %w", err)
	}

	// 1. Load image
	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if verbose {
		fmt.Fprintf(os.Stderr, "→ inspecting image %s\n", imageName)
	}

	img, err := image.Load(imageName, offline)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// check cache
	var digest string
	if img.Config != nil {
		digest = img.Config.RootFS.Type
		// try to use a better digest if available
		if img.Manifest != nil && len(img.Manifest.Layers) > 0 {
			digest = img.Manifest.Layers[0].Digest.String()
		}
	}

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if cached, found := cache.Get(digest); found && useCache {
		if verbose {
			fmt.Fprintf(os.Stderr, "→ using cached result from %s\n", cached.CachedAt.Format(time.RFC3339))
		}
		// could use cached result here if implementing full cache reuse
		// for now just log that we found it
	}

	// 2. Extract facts
	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if verbose {
		fmt.Fprintf(os.Stderr, "→ extracting facts\n")
	}

	extracted, err := facts.Extract(img)
	if err != nil {
		return fmt.Errorf("failed to extract facts: %w", err)
	}

	// 3. Load policy
	pol, err := loadPolicy(policyPath)
	if err != nil {
		return fmt.Errorf("failed to load policy: %w", err)
	}

	//nolint:errcheck // writes to stderr, nothing useful to do on error
	if verbose {
		fmt.Fprintf(os.Stderr, "→ evaluating %d rules\n", len(pol.Rules))
	}

	// 4. Evaluate rules
	engine := rules.NewEngine()
	results := engine.Evaluate(extracted, pol)

	// cache result
	if useCache {
		if cached, cacheErr := image.ImageToCache(img); cacheErr == nil {
			_ = cache.Put(digest, cached) // best effort caching, ignore errors
		}
	}

	// 5. Generate report
	rep := report.New(imageName, extracted, results, pol)

	switch format {
	case "json":
		return rep.JSON(os.Stdout)
	case "sarif":
		return rep.SARIF(os.Stdout)
	default:
		return rep.Text(os.Stdout)
	}
}

// cache operations
func handleCacheClear(dir string) error {
	cache, err := image.NewCache(dir, true)
	if err != nil {
		return err
	}

	if err := cache.Clear(); err != nil {
		return err
	}

	fmt.Println("cache cleared")
	return nil
}

func handleCacheStats(dir string) error {
	cache, err := image.NewCache(dir, true)
	if err != nil {
		return err
	}

	count, size, err := cache.Stats()
	if err != nil {
		return err
	}

	fmt.Printf("Cache Statistics\n")
	fmt.Printf("  Entries: %d\n", count)
	fmt.Printf("  Size:    %d bytes (%.2f MB)\n", size, float64(size)/(1024*1024))
	return nil
}

// bundle operations
func handleBundleImport(path string) error {
	mgr, err := packs.NewBundleManager("")
	if err != nil {
		return err
	}

	bundle, err := mgr.Import(path)
	if err != nil {
		return err
	}

	fmt.Printf("Imported bundle: %s v%s\n", bundle.Name, bundle.Version)
	fmt.Printf("  Description: %s\n", bundle.Description)
	fmt.Printf("  Policies: %d\n", len(bundle.Policies))
	if len(bundle.Advisories) > 0 {
		fmt.Printf("  Advisories: %d\n", len(bundle.Advisories))
	}
	return nil
}

func handleBundleList() error {
	mgr, err := packs.NewBundleManager("")
	if err != nil {
		return err
	}

	bundles, err := mgr.List()
	if err != nil {
		return err
	}

	if len(bundles) == 0 {
		fmt.Println("No bundles installed")
		return nil
	}

	fmt.Printf("Installed Bundles:\n\n")
	for _, b := range bundles {
		fmt.Printf("  %s (v%s)\n", b.Name, b.Version)
		fmt.Printf("    %s\n", b.Description)
		fmt.Printf("    Policies: %d\n", len(b.Policies))
		fmt.Println()
	}
	return nil
}

func handleBundleExport(outputPath string) error {
	// export current policies directory as a bundle
	name := "custom-policies"
	version := "1.0.0"
	description := "Custom policy bundle"

	if err := packs.Export("policies", outputPath, name, version, description); err != nil {
		return err
	}

	fmt.Printf("Exported bundle to %s\n", outputPath)
	return nil
}

func loadPolicy(path string) (*policy.Policy, error) {
	if path != "" {
		return policy.LoadFromFile(path)
	}

	// try to find default policy in a few places
	candidates := []string{
		"policies/default.yaml",
		"/etc/boltguard/default.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "boltguard", "default.yaml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return policy.LoadFromFile(candidate)
		}
	}

	// fall back to embedded default
	return policy.LoadDefault()
}
