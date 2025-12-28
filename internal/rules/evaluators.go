package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yourusername/boltguard/internal/facts"
	"github.com/yourusername/boltguard/internal/policy"
)

// UserEvaluator checks user/root configuration
type UserEvaluator struct{}

func (e *UserEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
	allowRoot := r.GetConfigBool("allow_root")

	if f.RunsAsRoot && !allowRoot {
		return &Result{
			Passed:  false,
			Message: fmt.Sprintf("image runs as root (user=%s)", f.User),
		}, nil
	}

	return &Result{
		Passed:  true,
		Message: fmt.Sprintf("runs as user: %s", f.User),
	}, nil
}

// SizeEvaluator checks image size limits
type SizeEvaluator struct{}

func (e *SizeEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
	maxMB := r.GetConfigInt("max_mb")
	warnMB := r.GetConfigInt("warn_mb")

	sizeMB := int(f.SizeMB())

	if maxMB > 0 && sizeMB > maxMB {
		return &Result{
			Passed:  false,
			Message: fmt.Sprintf("image size %dMB exceeds maximum %dMB", sizeMB, maxMB),
		}, nil
	}

	if warnMB > 0 && sizeMB > warnMB {
		return &Result{
			Passed:  true, // not a failure, just a warning
			Message: fmt.Sprintf("image size %dMB exceeds warning threshold %dMB", sizeMB, warnMB),
		}, nil
	}

	return &Result{
		Passed:  true,
		Message: fmt.Sprintf("image size: %dMB", sizeMB),
	}, nil
}

// LabelEvaluator checks for required labels
type LabelEvaluator struct{}

func (e *LabelEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
	required := r.GetConfigStringSlice("required")

	var missing []string
	for _, label := range required {
		if !f.HasLabel(label) {
			missing = append(missing, label)
		}
	}

	if len(missing) > 0 {
		return &Result{
			Passed:  false,
			Message: fmt.Sprintf("missing required labels: %s", strings.Join(missing, ", ")),
		}, nil
	}

	return &Result{
		Passed:  true,
		Message: fmt.Sprintf("all required labels present (%d total labels)", len(f.Labels)),
	}, nil
}

// EnvEvaluator checks environment variables
type EnvEvaluator struct{}

func (e *EnvEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
	denyPatterns := r.GetConfigStringSlice("deny_patterns")

	var violations []string

	for _, pattern := range denyPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern %s: %w", pattern, err)
		}

		for _, env := range f.Env {
			// check both key and value
			if re.MatchString(env) {
				// don't leak the actual value in the message
				parts := strings.SplitN(env, "=", 2)
				key := parts[0]
				violations = append(violations, fmt.Sprintf("env var %s matches pattern %s", key, pattern))
			}
		}
	}

	if len(violations) > 0 {
		return &Result{
			Passed:  false,
			Message: fmt.Sprintf("found suspicious env vars: %s", strings.Join(violations, "; ")),
		}, nil
	}

	return &Result{
		Passed:  true,
		Message: fmt.Sprintf("no suspicious env vars detected (%d total)", len(f.Env)),
	}, nil
}

// BaseEvaluator checks base image
type BaseEvaluator struct{}

func (e *BaseEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
	allowedPrefixes := r.GetConfigStringSlice("allowed_prefixes")
	allowUnknown := r.GetConfigBool("allow_unknown")

	if f.BaseImage == "" || f.BaseImage == "unknown" {
		if allowUnknown {
			return &Result{
				Passed:  true,
				Message: "base image unknown (allowed by policy)",
			}, nil
		}
		return &Result{
			Passed:  false,
			Message: "could not determine base image",
		}, nil
	}

	// check if base matches any allowed prefix
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(strings.ToLower(f.BaseImage), strings.ToLower(prefix)) {
			return &Result{
				Passed:  true,
				Message: fmt.Sprintf("base image: %s", f.BaseImage),
			}, nil
		}
	}

	// not in allowlist
	if allowUnknown {
		return &Result{
			Passed:  true, // warn but don't fail
			Message: fmt.Sprintf("base image %s not in recommended list", f.BaseImage),
		}, nil
	}

	return &Result{
		Passed:  false,
		Message: fmt.Sprintf("base image %s not allowed", f.BaseImage),
	}, nil
}

// LayersEvaluator checks layer count
type LayersEvaluator struct{}

func (e *LayersEvaluator) Evaluate(f *facts.Facts, r *policy.Rule) (*Result, error) {
	maxLayers := r.GetConfigInt("max_layers")
	warnLayers := r.GetConfigInt("warn_layers")

	count := f.LayerCount

	if maxLayers > 0 && count > maxLayers {
		return &Result{
			Passed:  false,
			Message: fmt.Sprintf("layer count %d exceeds maximum %d", count, maxLayers),
		}, nil
	}

	if warnLayers > 0 && count > warnLayers {
		return &Result{
			Passed:  true, // warning, not failure
			Message: fmt.Sprintf("layer count %d exceeds warning threshold %d", count, warnLayers),
		}, nil
	}

	return &Result{
		Passed:  true,
		Message: fmt.Sprintf("layer count: %d", count),
	}, nil
}
