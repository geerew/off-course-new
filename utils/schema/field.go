package schema

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/geerew/off-course/utils"
	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Field represents an individual field in a model
type Field struct {
	// Name is the name of the field. This is the same as `StructField.Name`
	Name string

	// Type is the type of the field (ex. string, *string, int, ...)
	Type reflect.Type

	// IndirectType is the indirect type of the field. Populated when the field is a pointer to
	// a type. This is populated with the concrete type (ex *string -> string)
	IndirectType reflect.Type

	// StructField is the reflect field of the field
	StructField reflect.StructField

	// EmbeddedSchema is the schema of the embedded field. Populated when the field is an
	// embedded struct
	EmbeddedSchema *Schema

	// NewTypePool is the pool of new values for the field
	NewTypePool TypePool

	// Column is the column name of the field in the database
	Column string

	// NotNull is a flag that indicates if the field is not null
	NotNull bool

	// Alias is the alias of the field in the database
	Alias string

	// IsJoin is a flag that indicates if the field is a join field
	IsJoin bool

	// Join is the join information, set when IsJoin is true
	Join map[string]string

	// ValueOf is a callback that returns the field's actual value and if it is zero from the given
	// struct. The callback differs based on whether the field is a simple field in the root struct
	// or a field of an embedded struct
	ValueOf func(context.Context, reflect.Value) (value any, zero bool)

	// ReflectValueOf is a callback that returns the field's reflect value for a given struct. The
	// callback differs based on whether the field is a simple field in the root struct or a field
	// of an embedded struct
	//
	// The difference between ValueOf and ReflectValueOf is that ValueOf returns the actual value
	// of the field as an `any` while ReflectValueOf returns the value as a reflect.Value
	ReflectValueOf func(context.Context, reflect.Value) reflect.Value

	// Set is a callback that sets the field's value based on the input value. It handles the
	// scenarios where the type of the value to set is different from the field type, for example
	// setting an int value to a string field
	Set func(context.Context, reflect.Value, any) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseField parses a struct field and returns a field
func ParseField(sf reflect.StructField, cache *sync.Map) (*Field, error) {
	tagSetting := parseDbTagSetting(sf.Tag.Get("db"))

	field := &Field{
		Name:         sf.Name,
		Type:         sf.Type,
		IndirectType: sf.Type,
		StructField:  sf,
		Column:       tagSetting["COL"],
		NotNull:      utils.CheckTruth(tagSetting["NOT NULL"], tagSetting["NOTNULL"]),
		IsJoin:       utils.CheckTruth(tagSetting["JOIN"]),
	}

	if field.IsJoin {
		field.Join = map[string]string{
			"JOIN":   tagSetting["JOIN"],
			"TABLE":  tagSetting["TABLE"],
			"COLUMN": tagSetting["COL"],
			"ON":     tagSetting["ON"],
		}
	}

	// Set the alias of the field
	if alias, ok := tagSetting["ALIAS"]; ok {
		field.Alias = alias
	}

	// Skip fields that are not db columns or embedded
	if field.Column == "" && !utils.CheckTruth(tagSetting["EMBEDDED"]) {
		return nil, nil
	}

	// While the field is a pointer, set the indirect type to the element type. This gets the
	// concrete type of the field
	for field.IndirectType.Kind() == reflect.Ptr {
		field.IndirectType = field.IndirectType.Elem()
	}

	// Create a working value of the field. This will be a pointer to the field type that
	// the data will be set to (ex *string, *int, ...)
	fieldValue := reflect.New(field.IndirectType)

	// When the field implements the driver.Valuer interface, set the field value to the value
	// returned by the Valuer.Value method. If it is a sql.Null* type, get the concrete value
	// of the field
	valuer, isValuer := fieldValue.Interface().(driver.Valuer)
	if isValuer {
		// If the value is valid and there is no error, set the field value to the value
		if v, err := valuer.Value(); reflect.ValueOf(v).IsValid() && err == nil {
			fieldValue = reflect.ValueOf(v)
		}

		// Get the first concrete value of the struct. This is used to get the value of sql.Null*
		// types
		fieldValue = getRealFieldValue(fieldValue)
	}

	// Handle fields that are structs but do not implement the driver.Valuer interface. Usually these
	// are embedded structs but could also be an unsupported type, like a map or slice
	if !isValuer {
		kind := reflect.Indirect(fieldValue).Kind()

		switch kind {
		case reflect.Struct:
			// If the field is a struct, parse the embedded schema
			var err error
			if field.EmbeddedSchema, _, err = validateAndParseFields(fieldValue.Interface(), cache); err != nil {
				return nil, err
			}

			// Prefix the position of the embedded schema in the root schema to each field in the
			// embedded schema. We use this to get the value of the embedded field in the likes of
			// ValueOf and ReflectValueOf.
			//
			// 1. When the embedded schema is a simple struct and at position 2 in the root schema,
			// the index of each field in the embedded schema will be prefixed with the positive
			// integer 2 (ex, [2, 0], [2, 1], ...)
			//
			// 2. When the embedded schema is NOT a simple struct, for example a pointer struct, and
			// is at position 2 in the root schema, the index of each field in the embedded schema
			// will be prefixed with the negative integer -2 (ex. [-2, 0], [-2, 1], ...)
			//
			// We can then use this later to traverse the struct fields using the index to get the value
			// of the field in the struct. If the index is negative, we know it is a pointer struct and
			// we can do extra work to dereference the pointer to get the value of the field
			for _, ef := range field.EmbeddedSchema.Fields {
				if field.Type.Kind() == reflect.Struct {
					ef.StructField.Index = append([]int{sf.Index[0]}, ef.StructField.Index...)
				} else {
					ef.StructField.Index = append([]int{-sf.Index[0] - 1}, ef.StructField.Index...)
				}
			}

		case reflect.Invalid, reflect.Uintptr, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface,
			reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
			return nil, fmt.Errorf("invalid embedded struct for field %s, should be struct, but got %v", field.Name, field.Type)
		}
	}

	return field, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (f *Field) setCallbacks() {
	f.NewTypePool = poolInitializer(reflect.PointerTo(f.IndirectType))
	f.setValueOfFn()
	f.setReflectValueOfFn()
	f.setSetFn()

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setValueOfFn sets the ValueOf callback for the field. The callback is used to get the field's
// actual value and if it is zero from the given struct. The callback differs based on whether the
// field is a simple field in the root struct or a field of an embedded struct
//
// When called, a struct is passed in and the reflect value of the field within that struct is
// returned
func (f *Field) setValueOfFn() {
	fieldIndex := f.StructField.Index[0]
	fieldIndexLen := len(f.StructField.Index)

	switch {
	case fieldIndexLen == 1:
		f.ValueOf = func(ctx context.Context, v reflect.Value) (any, bool) {
			fieldValue := reflect.Indirect(v).Field(fieldIndex)
			return fieldValue.Interface(), fieldValue.IsZero()
		}
	default:
		f.ValueOf = func(ctx context.Context, v reflect.Value) (any, bool) {
			v = reflect.Indirect(v)

			for _, fieldIdx := range f.StructField.Index {
				if fieldIdx >= 0 {
					v = v.Field(fieldIdx)
				} else {
					// Invert the negative index to get the index of the field in the struct
					v = v.Field(-fieldIdx - 1)

					if !v.IsNil() {
						v = v.Elem()
					} else {
						return nil, true
					}
				}
			}

			return v.Interface(), v.IsZero()
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setReflectValueOfFn sets the ReflectValueOf callback for the field. The callback is used to get
// the field's reflect value from the given struct. The callback differs based on whether the field
// is a simple field in the root struct or a field of an embedded struct
//
// When `ReflectValueOf` is called, a struct is passed in and the reflect.Value of the field within
// that struct is returned
//
// The difference between ValueOf and ReflectValueOf is that ValueOf returns the actual value
// of the field as an `any` while ReflectValueOf returns the value as a reflect.Value
func (f *Field) setReflectValueOfFn() {
	fieldIndex := f.StructField.Index[0]
	fieldIndexLen := len(f.StructField.Index)

	switch {
	case fieldIndexLen == 1:
		f.ReflectValueOf = func(ctx context.Context, v reflect.Value) reflect.Value {
			return reflect.Indirect(v).Field(fieldIndex)
		}
	default:
		f.ReflectValueOf = func(ctx context.Context, v reflect.Value) reflect.Value {
			v = reflect.Indirect(v)

			for idx, fieldIdx := range f.StructField.Index {
				if fieldIdx >= 0 {
					v = v.Field(fieldIdx)
				} else {
					// Invert the negative index to get the index of the field in the struct
					v = v.Field(-fieldIdx - 1)

					if v.IsNil() {
						v.Set(reflect.New(v.Type().Elem()))
					}

					if idx < len(f.StructField.Index)-1 {
						v = v.Elem()
					}
				}
			}

			return v
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// setSetFn sets the Set callback for the field. The callback is used to set the field's value
// based on the input value.
//
// It handles the scenarios where the type of the value to set is different from the field type,
// for example setting an int value to a string field
func (f *Field) setSetFn() {
	switch f.Type.Kind() {
	case reflect.Bool:
		f.Set = func(ctx context.Context, st reflect.Value, newValue interface{}) error {
			switch newData := newValue.(type) {
			case bool:
				f.ReflectValueOf(ctx, st).SetBool(newData)
			case int64:
				f.ReflectValueOf(ctx, st).SetBool(newData > 0)
			case string:
				b, _ := strconv.ParseBool(newData)
				f.ReflectValueOf(ctx, st).SetBool(b)
			default:
				return defaultSetter(ctx, st, newValue, f)
			}
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.Set = func(ctx context.Context, st reflect.Value, newValue interface{}) (err error) {
			switch newData := newValue.(type) {
			case int64:
				f.ReflectValueOf(ctx, st).SetInt(newData)
			case int, int8, int16, int32, uint, uint8, uint16, uint32, uint64, float32, float64:
				f.ReflectValueOf(ctx, st).SetInt(cast.ToInt64(newData))
			case []byte:
				return f.Set(ctx, st, string(newData))
			case string:
				if i, err := strconv.ParseInt(newData, 0, 64); err == nil {
					f.ReflectValueOf(ctx, st).SetInt(i)
				} else {
					return err
				}
			default:
				return defaultSetter(ctx, st, newValue, f)
			}
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f.Set = func(ctx context.Context, st reflect.Value, newValue interface{}) (err error) {
			switch newData := newValue.(type) {
			case uint64:
				f.ReflectValueOf(ctx, st).SetUint(newData)
			case uint, uint8, uint16, uint32, float32, float64:
				f.ReflectValueOf(ctx, st).SetUint(cast.ToUint64(newData))
			case []byte:
				return f.Set(ctx, st, string(newData))
			case string:
				if i, err := strconv.ParseUint(newData, 0, 64); err == nil {
					f.ReflectValueOf(ctx, st).SetUint(i)
				} else {
					return err
				}
			default:
				return defaultSetter(ctx, st, newValue, f)
			}
			return err
		}
	case reflect.Float32, reflect.Float64:
		f.Set = func(ctx context.Context, st reflect.Value, newValue interface{}) (err error) {
			switch newData := newValue.(type) {
			case float64:
				f.ReflectValueOf(ctx, st).SetFloat(newData)

			case float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				f.ReflectValueOf(ctx, st).SetFloat(cast.ToFloat64(newData))
			case []byte:
				return f.Set(ctx, st, string(newData))
			case string:
				if i, err := strconv.ParseFloat(newData, 64); err == nil {
					f.ReflectValueOf(ctx, st).SetFloat(i)
				} else {
					return err
				}
			default:
				return defaultSetter(ctx, st, newValue, f)
			}
			return err
		}
	case reflect.String:
		f.Set = func(ctx context.Context, st reflect.Value, newValue any) error {
			switch newData := newValue.(type) {
			case string:
				f.ReflectValueOf(ctx, st).SetString(newData)
			case []byte:
				f.ReflectValueOf(ctx, st).SetString(string(newData))
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				f.ReflectValueOf(ctx, st).SetString(cast.ToString(newData))
			case float64, float32:
				f.ReflectValueOf(ctx, st).SetString(fmt.Sprintf("%.f", newData))
			default:
				return defaultSetter(ctx, st, newValue, f)
			}

			return nil
		}
	default:
		workingValue := reflect.New(f.Type)

		if _, ok := workingValue.Elem().Interface().(sql.Scanner); ok {
			f.Set = func(ctx context.Context, st reflect.Value, newValue interface{}) (err error) {
				return structScannerSetter(ctx, st, newValue, f, true)
			}
		} else if _, ok := workingValue.Interface().(sql.Scanner); ok {
			f.Set = func(ctx context.Context, st reflect.Value, newValue interface{}) error {
				return structScannerSetter(ctx, st, newValue, f, false)
			}
		} else {
			f.Set = func(ctx context.Context, st reflect.Value, newValue any) error {
				return defaultSetter(ctx, st, newValue, f)
			}
		}

	}

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getRealFieldValue traverses a struct's fields to find the first concrete value of a field that is
// not a pointer. This is used to get the value of sql.Null* types
func getRealFieldValue(v reflect.Value) reflect.Value {
	v = reflect.Indirect(v)
	t := v.Type()

	// Check if the type is a struct and is not convertible to time.Time
	if t.Kind() == reflect.Struct && !t.ConvertibleTo(TimeReflectType) {
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i).Type

			// Unwrap the pointer type to get to the underlying field type
			for ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}

			// Recursively get the real field value
			fv := reflect.New(ft)
			if t != reflect.Indirect(fv).Type() {
				fv = getRealFieldValue(fv)
			}

			if fv.IsValid() {
				return fv
			}
		}
	}

	return v
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// defaultSetter is the default (fallback) setter function for a field
func defaultSetter(ctx context.Context, st reflect.Value, newValue any, f *Field) (err error) {
	if newValue == nil {
		// The new value is nil, set the field to the zero value based on the field type
		f.ReflectValueOf(ctx, st).Set(reflect.New(f.Type).Elem())
		return nil
	}

	newValueOf := reflect.ValueOf(newValue)
	newType := newValueOf.Type()

	// The new value is directly assignable to the field
	if newType.AssignableTo(f.Type) {
		if newValueOf.Kind() == reflect.Ptr && newValueOf.Elem().Kind() == reflect.Ptr {
			newValueOf = reflect.Indirect(newValueOf)
		}

		f.ReflectValueOf(ctx, st).Set(newValueOf)
		return
	}

	// The new value is convertible to the field type
	if newType.ConvertibleTo(f.Type) {
		f.ReflectValueOf(ctx, st).Set(newValueOf.Convert(f.Type))
		return
	}

	// When the field is a pointer, reflect on the field and determine if the new value
	// is assignable or convertible to the field type
	if f.Type.Kind() == reflect.Ptr {
		existingValueOf := f.ReflectValueOf(ctx, st)
		existingType := f.Type.Elem()

		if newType.AssignableTo(existingType) {
			if !existingValueOf.IsValid() {
				existingValueOf = reflect.New(existingType)
			} else if existingValueOf.IsNil() {
				existingValueOf.Set(reflect.New(existingType))
			}
			existingValueOf.Elem().Set(newValueOf)
			return
		} else if newType.ConvertibleTo(existingType) {
			if existingValueOf.IsNil() {
				existingValueOf.Set(reflect.New(existingType))
			}

			// Handle custom casting using the cast package for base types
			switch existingType.Kind() {
			case reflect.String:
				castedValue := cast.ToString(newValue)
				existingValueOf.Elem().Set(reflect.ValueOf(castedValue))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				castedValue := cast.ToInt64(newValue)
				existingValueOf.Elem().Set(reflect.ValueOf(castedValue).Convert(existingType))
			case reflect.Float32, reflect.Float64:
				castedValue := cast.ToFloat64(newValue)
				existingValueOf.Elem().Set(reflect.ValueOf(castedValue).Convert(existingType))
			case reflect.Bool:
				castedValue := cast.ToBool(newValue)
				existingValueOf.Elem().Set(reflect.ValueOf(castedValue))
			default:
				existingValueOf.Elem().Set(newValueOf.Convert(existingType))
			}
			return
		}
	}

	// When the new value is a pointer, set the field to the value of the pointer
	//
	// It handles 3 scenarios:
	//  1. The new value is a pointer to nil - Set the field to nil
	//  2. The new value is a pointer to a value that is assignable to the field - Set the field
	//     to the value
	//  3. The new value is a pointer to a value that is not assignable to the field - Recall the
	//     the fields Set callback with the value of the pointer
	if newValueOf.Kind() == reflect.Ptr {
		if newValueOf.IsNil() {
			f.ReflectValueOf(ctx, st).Set(reflect.New(f.Type).Elem())
		} else if newValueOf.Type().Elem().AssignableTo(f.Type) {
			f.ReflectValueOf(ctx, st).Set(newValueOf.Elem())
		} else {
			err = f.Set(ctx, st, newValueOf.Elem().Interface())
		}

		return
	}

	// When the new value implements the driver.Valuer interface, get the concrete value
	// then call the fields Set() callback
	if valuer, ok := newValue.(driver.Valuer); ok {
		if newValue, err = valuer.Value(); err == nil {
			err = f.Set(ctx, st, newValue)
			return
		}
	}

	return fmt.Errorf("unsupported type %T for field %s", newValue, f.Name)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// structScannerSetter is the setter function for fields that implement the sql.Scanner interface
func structScannerSetter(ctx context.Context, st reflect.Value, newValue any, f *Field, isPointer bool) error {
	newValueOf := reflect.ValueOf(newValue)

	// When the new value is not valid, set it to an new empty instance of the field
	// type
	if !newValueOf.IsValid() {
		f.ReflectValueOf(ctx, st).Set(reflect.New(f.Type).Elem())
		return nil
	}

	// When the new value is a pointer, call the fields Set callback with the value of
	// the pointer. If the pointer is nil, do nothing
	if newValueOf.Kind() == reflect.Ptr {
		if newValueOf.IsNil() {
			return nil
		}

		return f.Set(ctx, st, newValueOf.Elem().Interface())
	}

	// When the new value is simply assignable to the field type, set it
	if newValueOf.Type().AssignableTo(f.Type) {
		f.ReflectValueOf(ctx, st).Set(newValueOf)
		return nil
	}

	// When the field itself is nil, set an empty instance of the field type
	fieldValue := f.ReflectValueOf(ctx, st)
	if fieldValue.IsNil() {
		fieldValue.Set(reflect.New(f.Type.Elem()))
	}

	// When the new value implements the driver.Valuer interface, get the concrete value
	if valuer, ok := newValue.(driver.Valuer); ok {
		newValue, _ = valuer.Value()
	}

	// If this is already a pointer, just get the underlying value. If it is not a pointer, get
	// the address to the underlying value
	var s any
	if isPointer {
		s = fieldValue.Interface()
	} else {
		s = fieldValue.Addr().Interface()
	}

	return s.(sql.Scanner).Scan(newValue)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseDbTagSetting parses the `db` tag setting string and returns a map of the settings
func parseDbTagSetting(str string) map[string]string {
	settings := map[string]string{}

	sep := ";"
	names := strings.Split(str, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}
