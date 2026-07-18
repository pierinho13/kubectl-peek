package usage

type Relation string

const (
	RelationUses       Relation = "uses"
	RelationProduces   Relation = "produces"
	RelationReferences Relation = "references"
)

type Source string

const (
	SourceBuiltIn Source = "built-in"
	SourceRule    Source = "rule"
)

type Reference struct {
	Description string
	Path        string
	Key         string
	Relation    Relation
}

type Usage struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Source     Source
	References []Reference
}

type Warning struct {
	APIVersion string
	Resource   string
	Err        error
}

type Result struct {
	Usages   []Usage
	Warnings []Warning
}
