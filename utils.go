package main

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/idna"
)

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

func SplitEmail(email string) (string, string) {
	parts := strings.Split(email, "@")
	domain := parts[len(parts)-1]
	user := parts[0]

	san_domain, _ := idna.ToASCII(domain)
	return user, san_domain
}

func SanitizeEmail(email string) string {
	user, domain := SplitEmail(email)
	return user + "@" + domain
}

func JsonAny2StringList(input any) ([]string, bool) {
	fully_converted := true
	ret := []string{}

	if input == nil {
		return ret, false
	}

	if elements, ok := input.([]any); ok {
		for _, element := range elements {
			if str_element, ok := element.(string); ok {
				ret = append(ret, str_element)
			} else {
				ret = append(ret, fmt.Sprint(element))
			}
		}
	} else {
		return ret, false
	}

	return ret, fully_converted
}

func AssertError(err error) {
	if err != nil {
		panic(err)
	}
}

func TruncateText(input string, limit int) string {
	buff := bytes.NewBufferString("")

	for i, c := range input {
		if i >= limit {
			return buff.String()
		}
		buff.WriteRune(c)
	}

	return buff.String()
}

func Ternary[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}
	return b
}
