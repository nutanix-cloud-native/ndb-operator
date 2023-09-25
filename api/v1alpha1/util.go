package v1alpha1

import "sort"

func sortKeys(dict map[string]bool) []string {
	var keys []string
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
