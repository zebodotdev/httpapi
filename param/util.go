package param

import (
	"fmt"
	"strings"
)

func zeroValue[T any]() T {
	var value T
	return value
}

func joinPath(parent string, name string) string {
	parent = strings.TrimSpace(parent)
	name = strings.TrimSpace(name)
	if parent == "" {
		return name
	}
	if name == "" {
		return parent
	}
	return parent + "." + name
}

func normalizeNames(label string, names []string) []string {
	normalized := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			panic(fmt.Sprintf("httpapi/param: %s name cannot be empty", label))
		}
		key := strings.ToLower(name)
		if seen[key] {
			panic(fmt.Sprintf("httpapi/param: duplicate %s name %q", label, name))
		}
		seen[key] = true
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		panic(fmt.Sprintf("httpapi/param: at least one %s name is required", label))
	}
	return normalized
}
