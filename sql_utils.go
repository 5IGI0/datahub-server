package main

import "bytes"

func SQLEscapeStringLike(str string) string {
	result := bytes.NewBufferString("")

	for _, r := range str {
		if r == '%' || r == '_' || r == '\\' {
			result.WriteByte('\\')
		}
		result.WriteRune(r)
	}

	return result.String()
}
