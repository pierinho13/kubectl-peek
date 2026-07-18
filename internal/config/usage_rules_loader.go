package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

type UsageRulesFile struct {
	Rules []UsageRule `yaml:"rules"`
}

func LoadUsageRules(path string) ([]UsageRule, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read usage rules file %q: %w",
			path,
			err,
		)
	}

	var rulesFile UsageRulesFile

	if err := yaml.UnmarshalStrict(data, &rulesFile); err != nil {
		return nil, fmt.Errorf(
			"parse usage rules file %q: %w",
			path,
			err,
		)
	}

	if len(rulesFile.Rules) == 0 {
		return nil, fmt.Errorf(
			"usage rules file %q contains no rules",
			path,
		)
	}

	return rulesFile.Rules, nil
}
