package config

type Relation string

const (
	RelationUses       Relation = "uses"
	RelationProduces   Relation = "produces"
	RelationReferences Relation = "references"
)

type UsageRule struct {
	APIVersions []string              `yaml:"apiVersions"`
	Kind        string                `yaml:"kind"`
	Resource    string                `yaml:"resource"`
	References  []SecretReferenceRule `yaml:"references"`
}

type SecretReferenceRule struct {
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Relation    Relation `yaml:"relation"`
}
