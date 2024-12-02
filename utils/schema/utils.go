package schema

import (
	"reflect"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseRelatedSchema parses the related schema and returns a reflect value that can
// be used to set the related field
func parseRelatedSchema(rel *relation) (*Schema, reflect.Value, error) {
	concreteType := rel.RelatedType
	for concreteType.Kind() == reflect.Ptr {
		concreteType = concreteType.Elem()
	}

	relatedModelPtr := reflect.New(concreteType)
	relatedModel := relatedModelPtr.Interface()

	relatedSchema, err := Parse(relatedModel)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	return relatedSchema, relatedModelPtr, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getStructField returns a struct field at the given position
func getStructField(concreteRv reflect.Value, position []int) reflect.Value {
	structField := concreteRv
	for _, pos := range position {
		structField = reflect.Indirect(structField).Field(pos)
	}

	return structField
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setRelatedField sets the related field
func setRelatedField(relatedField reflect.Value, value reflect.Value) {
	if relatedField.Kind() == reflect.Ptr {
		if value.Kind() == reflect.Ptr {
			relatedField.Set(value)
		} else {
			relatedField.Set(value.Addr())
		}
	} else {
		if value.Kind() == reflect.Ptr {
			relatedField.Set(value.Elem())
		} else {
			relatedField.Set(value)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// concreteReflectValue returns the concrete reflect value
func concreteReflectValue(v reflect.Value) (reflect.Value, error) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() && v.CanAddr() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	if !v.IsValid() {
		return v, utils.ErrInvalidValue
	}

	return v, nil
}
