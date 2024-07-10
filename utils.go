package main

import "bytes"

func reverse_str(str string) string {
	runes := make([]rune, 0, len(str))

	for _, r := range str {
		runes = append(runes, r)
	}

	result := bytes.NewBufferString("")

	for i := 0; i < len(runes); i++ {
		result.WriteRune(runes[len(runes)-i-1])
	}

	return result.String()
}

func alnumify(str string) string {
	result := bytes.NewBufferString("")

	for _, r := range str {
		if (r <= 'z' && r >= 'a') ||
			(r <= 'Z' && r >= 'A') ||
			(r <= '9' && r >= '0') {
			result.WriteRune(r)
		}
	}

	return result.String()
}
