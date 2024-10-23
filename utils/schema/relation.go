package schema

import (
	"reflect"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type relation struct {
	// The name of the struct field
	Name string

	// The position of the field in the struct
	Position []int

	// True when the relation is a has many (i.e. a slice)
	HasMany bool

	// The field to match on in the relation
	MatchOn string

	// The type of the related struct
	RelatedType reflect.Type

	// When true, the relation is a pointer
	IsPtr bool
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseField will parse a field from the struct and return a field struct
func parseRelation(sf reflect.StructField, config *modelRelationConfig) *relation {
	r := &relation{
		Name:        sf.Name,
		Position:    sf.Index,
		MatchOn:     config.match,
		RelatedType: sf.Type,
		IsPtr:       sf.Type.Kind() == reflect.Ptr,
	}

	concreteType := sf.Type
	for concreteType.Kind() == reflect.Ptr {
		concreteType = concreteType.Elem()
	}

	if concreteType.Kind() == reflect.Slice {
		r.HasMany = true
		r.RelatedType = concreteType.Elem()
	}

	return r
}
