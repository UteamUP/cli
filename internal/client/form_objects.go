package client

import (
	"fmt"
	"strings"
)

// expandFormObjects preserves explicit nulls and allows only one level of
// scalar fields for controller models that opt into nested form binding.
func expandFormObjects(params map[string]any, objectFields []string) (map[string]any, error) {
	allowed := make(map[string]bool, len(objectFields))
	for _, field := range objectFields {
		allowed[strings.ToLower(field)] = true
	}
	fields := make(map[string]any, len(params))
	seen := make(map[string]bool, len(params))
	add := func(key string, value any) error {
		identity := strings.ToLower(key)
		if seen[identity] {
			return fmt.Errorf("ambiguous form field %q", key)
		}
		seen[identity] = true
		fields[key] = value
		return nil
	}
	for key, value := range params {
		if key == "" || strings.ContainsAny(key, ".[]") {
			return nil, fmt.Errorf("field %q must use the reviewed nested JSON object", key)
		}
		object, isObject := value.(map[string]any)
		if !isObject || !allowed[strings.ToLower(key)] {
			if err := add(key, value); err != nil {
				return nil, err
			}
			continue
		}
		// Reserve the parent name too: Mutation and mutation cannot shadow
		// each other under the case-insensitive ASP.NET form binder.
		if err := add(key, nil); err != nil {
			return nil, err
		}
		for child, childValue := range object {
			if child == "" || strings.ContainsAny(child, ".[]") {
				return nil, fmt.Errorf("invalid nested form field %q", child)
			}
			name := key + "." + child
			if childValue != nil {
				if _, err := formScalar(name, childValue); err != nil {
					return nil, err
				}
			}
			if err := add(name, childValue); err != nil {
				return nil, err
			}
		}
	}
	return fields, nil
}
