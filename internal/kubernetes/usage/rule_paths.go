package usage

import (
	"fmt"
	"strings"
)

const wildcardSuffix = "[*]"

type rulePathSegment struct {
	field    string
	wildcard bool
}

func findStringValuesAtPath(
	object map[string]interface{},
	path string,
) ([]string, error) {
	segments, err := parseRulePath(path)
	if err != nil {
		return nil, err
	}

	currentValues := []interface{}{
		object,
	}

	for _, segment := range segments {
		nextValues := make([]interface{}, 0)

		for _, currentValue := range currentValues {
			objectValue, ok := currentValue.(map[string]interface{})
			if !ok {
				continue
			}

			fieldValue, found := objectValue[segment.field]
			if !found {
				continue
			}

			if !segment.wildcard {
				nextValues = append(
					nextValues,
					fieldValue,
				)

				continue
			}

			arrayValue, ok := fieldValue.([]interface{})
			if !ok {
				continue
			}

			nextValues = append(
				nextValues,
				arrayValue...,
			)
		}

		currentValues = nextValues

		if len(currentValues) == 0 {
			return nil, nil
		}
	}

	values := make([]string, 0, len(currentValues))

	for _, currentValue := range currentValues {
		stringValue, ok := currentValue.(string)
		if !ok || stringValue == "" {
			continue
		}

		values = append(
			values,
			stringValue,
		)
	}

	return values, nil
}

func parseRulePath(
	path string,
) ([]rulePathSegment, error) {
	path = strings.TrimSpace(path)

	if path == "" {
		return nil, fmt.Errorf("rule path is empty")
	}

	rawSegments := strings.Split(path, ".")

	segments := make(
		[]rulePathSegment,
		0,
		len(rawSegments),
	)

	for _, rawSegment := range rawSegments {
		rawSegment = strings.TrimSpace(rawSegment)

		if rawSegment == "" {
			return nil, fmt.Errorf(
				"invalid empty segment in rule path %q",
				path,
			)
		}

		segment := rulePathSegment{
			field: rawSegment,
		}

		if strings.HasSuffix(
			rawSegment,
			wildcardSuffix,
		) {
			segment.field = strings.TrimSuffix(
				rawSegment,
				wildcardSuffix,
			)
			segment.wildcard = true
		}

		if segment.field == "" {
			return nil, fmt.Errorf(
				"invalid wildcard segment in rule path %q",
				path,
			)
		}

		if strings.Contains(segment.field, "[") ||
			strings.Contains(segment.field, "]") {
			return nil, fmt.Errorf(
				"unsupported array expression in rule path %q",
				path,
			)
		}

		segments = append(
			segments,
			segment,
		)
	}

	return segments, nil
}
