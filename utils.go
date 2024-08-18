package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

func SanitizeDomain(domain string) (string, bool) {
	port_splitted := strings.Split(domain, ":")

	if len(port_splitted) == 2 {
		// domain with port
		domain = port_splitted[0]
	} else if len(port_splitted) > 2 {
		// might be some IPv6 or non-valid domain
		return "", false
	}

	if len(domain) == 0 {
		return "", false
	}

	domain = strings.ToLower(domain)

	c := domain[len(domain)-1]
	if c <= '9' && c >= '0' {
		// can't tell where i read it
		// but i know domains can't end with numbers
		// TODO: check TLD list
		return "", false
	}

	domain, err := idna.ToASCII(domain)
	if err != nil {
		return "", false
	}

	domain = strings.TrimRight(domain, ".")

	return domain, strings.ContainsRune(domain, '.')
}

func SanitizeDomains(domains []string) []string {
	ret := make([]string, 0, len(domains))
	for _, domain := range domains {
		if san_domain, e := SanitizeDomain(domain); e {
			ret = append(ret, san_domain)
		}
	}

	return ret
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

func Assert(cond bool) {
	if !cond {
		panic(errors.New("failed assert"))
	}
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

func Req2Page(r *http.Request) (int, int) {
	page_size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	page, _ := strconv.Atoi(mux.Vars(r)["page"])

	if page <= 0 {
		page = 1
	}

	if page_size <= 0 {
		page_size = DEFAULT_PAGESIZE
	}

	if page_size > MAX_PAGESIZE {
		page_size = MAX_PAGESIZE
	}

	return page, page_size
}

func ForceInt64Cast(untyped any) (int64, bool) {
	switch val := untyped.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case float64:
		return int64(val), true
	case float32:
		return int64(val), true

	}

	return 0, false
}

func ExtractDomainFromLink(link string) (string, bool) {

	url_parts := strings.Split(link, "/")
	if len(url_parts) < 3 {
		return "", false
	}

	return strings.Split(url_parts[2], ":")[0], true
}
