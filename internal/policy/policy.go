package policy

import (
	"embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed defaults/default.yaml
var defaultPolicyFS embed.FS

// Policy represents a complete policy document
type Policy struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`

	// Global settings
	Settings Settings `yaml:"settings"`

	// Rules to evaluate
	Rules []Rule `yaml:"rules"`
}

type Settings struct {
	// FailOnError - if true, any error during evaluation fails the whole check
	FailOnError bool `yaml:"fail_on_error"`

	// Severity threshold - only report issues at this level or higher
	MinSeverity string `yaml:"min_severity"`
}

// Rule defines a single check to perform
type Rule struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Severity    string `yaml:"severity"` // critical, high, medium, low, info
	Kind        string `yaml:"kind"`     // user, size, label, env, etc.

	// rule-specific config as free-form map
	Config map[string]interface{} `yaml:"config"`

	// optional - if this rule fails, should we fail the whole check?
	FailFast bool `yaml:"fail_fast"`
}

// LoadFromFile loads a policy from a YAML file
func LoadFromFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse policy YAML: %w", err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid policy: %w", err)
	}

	return &p, nil
}

// LoadDefault loads the embedded default policy
func LoadDefault() (*Policy, error) {
	data, err := defaultPolicyFS.ReadFile("defaults/default.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded default policy: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse default policy: %w", err)
	}

	return &p, nil
}

// Validate checks if the policy is well-formed
func (p *Policy) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("policy must have a name")
	}

	if len(p.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}

	// check each rule
	for i, rule := range p.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule %d missing ID", i)
		}
		if rule.Name == "" {
			return fmt.Errorf("rule %s missing name", rule.ID)
		}
		if rule.Kind == "" {
			return fmt.Errorf("rule %s missing kind", rule.ID)
		}

		// validate severity
		switch rule.Severity {
		case "critical", "high", "medium", "low", "info":
			// valid
		default:
			return fmt.Errorf("rule %s has invalid severity: %s", rule.ID, rule.Severity)
		}
	}

	return nil
}

// GetConfigString retrieves a string value from rule config
func (r *Rule) GetConfigString(key string) string {
	if r.Config == nil {
		return ""
	}
	val, ok := r.Config[key]
	if !ok {
		return ""
	}
	str, _ := val.(string)
	return str
}

// GetConfigInt retrieves an int value from rule config
func (r *Rule) GetConfigInt(key string) int {
	if r.Config == nil {
		return 0
	}
	val, ok := r.Config[key]
	if !ok {
		return 0
	}

	// yaml can parse as int or float64
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// GetConfigBool retrieves a bool value from rule config
func (r *Rule) GetConfigBool(key string) bool {
	if r.Config == nil {
		return false
	}
	val, ok := r.Config[key]
	if !ok {
		return false
	}
	b, _ := val.(bool)
	return b
}

// GetConfigStringSlice retrieves a []string from rule config
func (r *Rule) GetConfigStringSlice(key string) []string {
	if r.Config == nil {
		return nil
	}
	val, ok := r.Config[key]
	if !ok {
		return nil
	}

	// could be []interface{} or []string
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}
