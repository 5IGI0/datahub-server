package main

import (
	"strings"

	"github.com/Masterminds/squirrel"
)

type SqlCondGenerator interface {
	Generate(field string, vals []string) ([]squirrel.Sqlizer, error)
}

type EqualCondGeneratorType struct{}

var EqualCondGenerator EqualCondGeneratorType

func (EqualCondGeneratorType) Generate(field string, vals []string) ([]squirrel.Sqlizer, error) {
	return []squirrel.Sqlizer{squirrel.Eq{field: vals[0]}}, nil
}

type BeginsWithCondGeneratorType struct{}

var BeginsWithCondGenerator BeginsWithCondGeneratorType

func (BeginsWithCondGeneratorType) Generate(field string, vals []string) ([]squirrel.Sqlizer, error) {
	return []squirrel.Sqlizer{squirrel.Like{
		field: SQLEscapeStringLike(vals[0]) + "%"}}, nil
}

type BoolCondGeneratorType struct{}

var BoolCondGenerator BoolCondGeneratorType

func (BoolCondGeneratorType) Generate(field string, vals []string) ([]squirrel.Sqlizer, error) {
	v := strings.ToLower(vals[0])
	vv := 0
	if v == "1" || v == "true" {
		vv = 1
	}

	return []squirrel.Sqlizer{squirrel.Eq{
		field: vv}}, nil
}

type SubDomainCondGeneratorType struct{}

var SubDomainCondGenerator SubDomainCondGeneratorType

func (SubDomainCondGeneratorType) Generate(field string, vals []string) ([]squirrel.Sqlizer, error) {
	return []squirrel.Sqlizer{
		squirrel.Expr(
			"`"+field+"` LIKE REVERSE(?)",
			"%"+SQLEscapeStringLike(vals[0]),
		)}, nil
}

type ToggleCondGeneratorType struct{}

var ToggleCondGenerator ToggleCondGeneratorType

func (ToggleCondGeneratorType) Generate(field string, vals []string) ([]squirrel.Sqlizer, error) {
	v := strings.ToLower(vals[0])
	if v == "1" || v == "true" {
		return []squirrel.Sqlizer{squirrel.Eq{
			field: 1}}, nil
	}
	return nil, nil
}

type LikeCondGeneratorType struct{}

var LikeCondGenerator LikeCondGeneratorType

func (LikeCondGeneratorType) Generate(field string, vals []string) ([]squirrel.Sqlizer, error) {
	return []squirrel.Sqlizer{squirrel.Like{
		field: SQLEscapeStringLike(vals[0])}}, nil
}
