package main

import (
	"bytes"
	"net/url"

	"github.com/Masterminds/squirrel"
)

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

type Query2SqlCond struct {
	Field     string
	Generator SqlCondGenerator
	Default   string
}

func GetQuery2SqlConds(vals url.Values, config map[string]Query2SqlCond) (squirrel.And, error) {
	var ret squirrel.And
	for k, v := range vals {
		if generator, e := config[k]; e {
			field := generator.Field
			if field == "" {
				field = k
			}

			tmp, err := generator.Generator.Generate(
				field, v)

			if err != nil {
				return nil, err
			}

			ret = append(ret, tmp...)
		}
	}

	/* default values */
	for k, v := range config {
		if _, e := vals[k]; !e {
			if v.Default == "" {
				continue
			}

			field := v.Field
			if field == "" {
				field = k
			}

			tmp, err := v.Generator.Generate(
				field, []string{v.Default})

			if err != nil {
				return nil, err
			}

			ret = append(ret, tmp...)
		}
	}

	return ret, nil
}
