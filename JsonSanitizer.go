package main

import (
	"fmt"
	"slices"
	"strings"
)

func JsonSanitize(input any) any {
	switch val := input.(type) {
	case []any:
		ret := make([]any, 0, len(val))

		for _, v := range val {
			ret = append(ret, JsonSanitize(v))
		}

		return JsonSortList(ret)
	case map[string]any:
		ret := make(map[string]any)
		for k, v := range val {
			ret[k] = JsonSanitize(v)
		}
		return ret
	default:
		return input
	}
}

// TODO: remove duplicates
func JsonSortList(list []any) []any {
	output := make([]any, len(list))
	copy(output, list)

	slices.SortFunc(output, func(a any, b any) int {
		return strings.Compare(
			fmt.Sprintf("%10T:%v", a, a),
			fmt.Sprintf("%10T:%v", b, b))
	})

	return output
}
