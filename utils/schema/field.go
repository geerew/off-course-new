package schema

import (
	"reflect"

	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type field struct {
	// The name of the struct field
	Name string

	// The position of the field in the struct
	Position []int

	// The name of the column in the database
	Column string

	// A db column alias
	Alias string

	// When true, the field cannot be null in the database
	NotNull bool

	// When true, the field can be updated
	Mutable bool

	// When true, the field will be skipped during create if it is null
	IgnoreIfNull bool

	// The name of the join table that this field is associated with
	JoinTable string

	// ReflectValueOf is a callback that takes a struct, as a reflect.Value, and returns the
	// reflect.Value of the field
	ReflectValueOf func(reflect.Value) reflect.Value

	// ValueOf is a callback that takes a struct, as a reflect.Value, and returns the actual
	// value of the field and whether it is zero
	ValueOf func(reflect.Value) (any, bool)

	// The concrete type of the field
	concreteType reflect.Type
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseField will parse a field from the struct and return a field struct
func parseField(sf reflect.StructField, config *modelFieldConfig) *field {
	f := &field{
		Name:         sf.Name,
		Position:     sf.Index,
		Column:       config.column,
		Alias:        config.alias,
		NotNull:      config.notNull,
		Mutable:      config.mutable,
		IgnoreIfNull: config.ignoreIfNull,
		JoinTable:    config.joinTable,
		concreteType: sf.Type,
	}

	if f.Column == "" {
		f.Column = utils.SnakeCase(sf.Name)
	}

	// Drill down to the concrete type
	for f.concreteType.Kind() == reflect.Ptr {
		f.concreteType = f.concreteType.Elem()
	}

	f.setReflectValueOf()
	f.setValueOf()

	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseEmbeddedField will validate an embedded field implements the Modeler interface and then
// parses the fields of the embedded struct
func parseEmbeddedField(sf reflect.StructField) ([]*field, error) {
	if sf.Type.Kind() != reflect.Struct {
		return nil, utils.ErrEmbedded
	}

	m, isDefiner := reflect.New(sf.Type).Interface().(Definer)
	if isDefiner {
		config := &ModelConfig{}
		m.Define(config)

		rt := reflect.Indirect(reflect.ValueOf(m)).Type()

		fields := []*field{}
		for i := range rt.NumField() {
			sf := rt.Field(i)

			for sf.Type.Kind() == reflect.Ptr {
				sf.Type = sf.Type.Elem()
			}

			if fieldConfig, ok := config.fields[sf.Name]; ok {
				fields = append(fields, parseField(sf, fieldConfig))
			} else if _, ok := config.embedded[sf.Name]; ok {
				fs, err := parseEmbeddedField(sf)
				if err != nil {
					return nil, err
				}

				fields = append(fields, fs...)
			}
		}

		for _, field := range fields {
			field.Position = append(sf.Index, field.Position...)
		}

		return fields, nil
	} else {
		return nil, utils.ErrEmbedded
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetReflectValueOf sets the ReflectValueOf callback for the field
func (f *field) setReflectValueOf() {
	f.ReflectValueOf = func(rv reflect.Value) reflect.Value {
		value := reflect.Indirect(rv)
		if len(f.Position) == 1 {
			value = value.Field(f.Position[0])
		} else {
			for _, p := range f.Position {
				value = value.Field(p)
			}
		}

		return value
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetValueOf sets the ValueOf callback for the field. The callback will return the actual value
// of the field from the given struct and if the value is zero
func (f *field) setValueOf() {
	f.ValueOf = func(rv reflect.Value) (any, bool) {
		value := f.ReflectValueOf(rv)
		return value.Interface(), value.IsZero()
	}
}
