package usage

import (
	"context"
	"fmt"
	"strings"

	"github.com/pierinho13/kubectl-peek/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type RuleFinder struct {
	discoveryClient discovery.DiscoveryInterface
	dynamicClient   dynamic.Interface
	rules           []config.UsageRule
}

func NewRuleFinder(
	discoveryClient discovery.DiscoveryInterface,
	dynamicClient dynamic.Interface,
	rules []config.UsageRule,
) *RuleFinder {
	return &RuleFinder{
		discoveryClient: discoveryClient,
		dynamicClient:   dynamicClient,
		rules:           rules,
	}
}

func (finder *RuleFinder) Find(
	ctx context.Context,
	namespace string,
	secretName string,
) Result {
	var result Result

	if finder == nil ||
		finder.discoveryClient == nil ||
		finder.dynamicClient == nil {
		return result
	}

	for _, rule := range finder.rules {
		finder.findRuleUsages(
			ctx,
			namespace,
			secretName,
			rule,
			&result,
		)
	}

	return result
}

func (finder *RuleFinder) findRuleUsages(
	ctx context.Context,
	namespace string,
	secretName string,
	rule config.UsageRule,
	result *Result,
) {
	gvrs := ruleGroupVersionResources(
		result,
		rule,
	)
	if len(gvrs) == 0 {
		return
	}

	validReferences := validateReferenceRules(
		result,
		rule,
	)
	if len(validReferences) == 0 {
		return
	}

	gvr, available, err := firstAvailableResource(
		finder.discoveryClient,
		gvrs,
	)
	if err != nil {
		appendRuleWarning(
			result,
			ruleAPIVersionsLabel(rule),
			rule.Resource,
			fmt.Errorf(
				"discover resource %q: %w",
				rule.Resource,
				err,
			),
		)

		return
	}

	if !available {
		return
	}

	objects, err := finder.dynamicClient.
		Resource(gvr).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendRuleWarning(
			result,
			gvr.GroupVersion().String(),
			rule.Resource,
			fmt.Errorf(
				"list %s: %w",
				gvr.String(),
				err,
			),
		)

		return
	}

	for i := range objects.Items {
		object := &objects.Items[i]

		references := findObjectRuleReferences(
			object.Object,
			secretName,
			validReferences,
		)
		if len(references) == 0 {
			continue
		}

		apiVersion := object.GetAPIVersion()
		if apiVersion == "" {
			apiVersion = gvr.GroupVersion().String()
		}

		kind := object.GetKind()
		if kind == "" {
			kind = rule.Kind
		}

		result.Usages = append(
			result.Usages,
			Usage{
				APIVersion: apiVersion,
				Kind:       kind,
				Namespace:  object.GetNamespace(),
				Name:       object.GetName(),
				Source:     SourceRule,
				References: references,
			},
		)
	}
}

func ruleGroupVersionResources(
	result *Result,
	rule config.UsageRule,
) []schema.GroupVersionResource {
	if rule.Resource == "" {
		appendRuleWarning(
			result,
			ruleAPIVersionsLabel(rule),
			"",
			fmt.Errorf(
				"resource is empty for kind %q",
				rule.Kind,
			),
		)

		return nil
	}

	if len(rule.APIVersions) == 0 {
		appendRuleWarning(
			result,
			"",
			rule.Resource,
			fmt.Errorf(
				"apiVersions is empty for resource %q",
				rule.Resource,
			),
		)

		return nil
	}

	resources := make(
		[]schema.GroupVersionResource,
		0,
		len(rule.APIVersions),
	)

	seen := make(map[string]struct{})

	for _, apiVersion := range rule.APIVersions {
		apiVersion = strings.TrimSpace(apiVersion)

		if apiVersion == "" {
			appendRuleWarning(
				result,
				"",
				rule.Resource,
				fmt.Errorf(
					"apiVersion is empty for resource %q",
					rule.Resource,
				),
			)

			continue
		}

		if _, found := seen[apiVersion]; found {
			continue
		}

		seen[apiVersion] = struct{}{}

		groupVersion, err := schema.ParseGroupVersion(
			apiVersion,
		)
		if err != nil {
			appendRuleWarning(
				result,
				apiVersion,
				rule.Resource,
				fmt.Errorf(
					"parse apiVersion %q: %w",
					apiVersion,
					err,
				),
			)

			continue
		}

		resources = append(
			resources,
			groupVersion.WithResource(rule.Resource),
		)
	}

	return resources
}

func validateReferenceRules(
	result *Result,
	rule config.UsageRule,
) []config.SecretReferenceRule {
	validReferences := make(
		[]config.SecretReferenceRule,
		0,
		len(rule.References),
	)

	apiVersions := ruleAPIVersionsLabel(rule)

	for _, referenceRule := range rule.References {
		if _, err := parseRulePath(referenceRule.Path); err != nil {
			appendRuleWarning(
				result,
				apiVersions,
				rule.Resource,
				fmt.Errorf(
					"invalid path %q for %s: %w",
					referenceRule.Path,
					rule.Kind,
					err,
				),
			)

			continue
		}

		if _, err := convertRuleRelation(
			referenceRule.Relation,
		); err != nil {
			appendRuleWarning(
				result,
				apiVersions,
				rule.Resource,
				fmt.Errorf(
					"invalid relation for path %q: %w",
					referenceRule.Path,
					err,
				),
			)

			continue
		}

		validReferences = append(
			validReferences,
			referenceRule,
		)
	}

	return validReferences
}

func findObjectRuleReferences(
	object map[string]interface{},
	secretName string,
	rules []config.SecretReferenceRule,
) []Reference {
	references := make([]Reference, 0)

	for _, rule := range rules {
		values, err := findStringValuesAtPath(
			object,
			rule.Path,
		)
		if err != nil {
			// La ruta ya fue validada antes de procesar los objetos.
			continue
		}

		if !containsString(values, secretName) {
			continue
		}

		relation, err := convertRuleRelation(
			rule.Relation,
		)
		if err != nil {
			continue
		}

		references = append(
			references,
			Reference{
				Description: rule.Description,
				Path:        rule.Path,
				Relation:    relation,
			},
		)
	}

	return references
}

func convertRuleRelation(
	relation config.Relation,
) (Relation, error) {
	switch relation {
	case config.RelationUses:
		return RelationUses, nil

	case config.RelationProduces:
		return RelationProduces, nil

	case config.RelationReferences:
		return RelationReferences, nil

	default:
		return "", fmt.Errorf(
			"unsupported relation %q",
			relation,
		)
	}
}

func containsString(
	values []string,
	expected string,
) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}

	return false
}

func ruleAPIVersionsLabel(
	rule config.UsageRule,
) string {
	return strings.Join(
		rule.APIVersions,
		",",
	)
}

func appendRuleWarning(
	result *Result,
	apiVersion string,
	resource string,
	err error,
) {
	result.Warnings = append(
		result.Warnings,
		Warning{
			APIVersion: apiVersion,
			Resource:   resource,
			Err:        err,
		},
	)
}
