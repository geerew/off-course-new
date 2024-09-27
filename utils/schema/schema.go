package schema

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var ErrUnsupportedType = errors.New("unsupported type")
var ErrNilType = errors.New("nil model")
var ErrInvalidValue = errors.New("invalid value")

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	TimeReflectType    = reflect.TypeOf(time.Time{})
	TimePtrReflectType = reflect.TypeOf(&time.Time{})
	ByteReflectType    = reflect.TypeOf(uint8(0))
	StringReflectType  = reflect.TypeOf("")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Schema represents a model schema
type Schema struct {
	// Name is the name of the struct, including the package name
	Name string

	// ReflectValue is the reflect value for the struct the schema represents
	ReflectValue reflect.Value

	// ReflectType is the reflect type for the struct the schema represents
	ReflectType reflect.Type

	// Fields is a list of fields in the schema
	Fields []*Field

	// FieldCols is a list of column names for the fields in the schema
	FieldCols []string

	// FieldsByName is a map of fields by their name, making it easier to access fields by name
	FieldsByName map[string]*Field

	// FieldsByColumn is a map of fields by their DB column name, making it easier to access fields
	// by column
	FieldsByColumn map[string]*Field
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Parse validates the input type is a struct and then parses each field of the struct including
// embedded structs. It then initializes object pools and setters for each field in the schema
func Parse(input any, cache *sync.Map) (*Schema, error) {
	if input == nil {
		return nil, ErrNilType
	}

	// Get the reflect value of the input and unwrap pointers
	reflectValue := reflect.ValueOf(input)
	for reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() && reflectValue.CanAddr() {
			reflectValue.Set(reflect.New(reflectValue.Type().Elem()))
		}

		reflectValue = reflectValue.Elem()
	}

	// Error if the reflect value is invalid. This can happen if the input is nil, or an
	// uninitialized pointer like `var test *Test`
	if !reflectValue.IsValid() {
		return nil, ErrInvalidValue
	}

	// Parse the fields of the input
	schema, cacheHit, err := validateAndParseFields(input, cache)
	if err != nil {
		return nil, err
	}

	// If the schema was loaded from cache, return it
	if cacheHit {
		return schema, nil
	}

	// Set the reflect value of the schema
	schema.ReflectValue = reflectValue

	// Slices to ease access to fields
	schema.FieldCols = make([]string, 0, len(schema.Fields))
	schema.FieldsByName = make(map[string]*Field, len(schema.Fields))
	schema.FieldsByColumn = make(map[string]*Field, len(schema.Fields))

	for _, f := range schema.Fields {
		schema.FieldCols = append(schema.FieldCols, f.Column)
		schema.FieldsByName[f.Name] = f
		schema.FieldsByColumn[f.Column] = f
		f.setCallbacks()
	}

	// Cache the schema, or load it if another goroutine has already cached it
	if v, loaded := cache.LoadOrStore(schema.ReflectType, schema); loaded {
		s := v.(*Schema)
		return s, nil
	}

	return schema, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// validateAndParseFields validates the input type and parses its fields to create a schema. It
// takes an input of any type and checks it is a struct or pointer to a struct. It uses reflection
// to handle different kinds of types, unwrapping pointers, slices, or arrays to get the concrete
// type. If the resulting type is not a struct, it returns an error
//
// Once the type is confirmed to be a struct, it parses each field of the struct, and for any embedded
// struct (schema), it recursively adds its fields to the final schema
func validateAndParseFields(input any, cache *sync.Map) (*Schema, bool, error) {
	if input == nil {
		return nil, false, ErrNilType
	}

	// Get the concrete value of the input. If it is a pointer and is nil, create a new instance
	// of the input
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}

	// Get the initial type of the input. Uses indirect to better handle pointers to the input
	inputType := reflect.Indirect(value).Type()

	// While the type is a slice, array, or pointer, set the type to the element type. This
	// gets the concrete type of the input
	for inputType.Kind() == reflect.Slice || inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}

	// If the type is not a struct, return an error
	if inputType.Kind() != reflect.Struct {
		return nil, false, fmt.Errorf("%w: %+v", ErrUnsupportedType, input)
	}

	// Attempt to load the schema from cache
	if v, ok := cache.Load(inputType); ok {
		s := v.(*Schema)
		return s, true, nil
	}

	schema := &Schema{
		Name:        inputType.Name(),
		ReflectType: inputType,
	}

	// Parse each field in the input
	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)

		if field, err := ParseField(field, cache); err != nil {
			return nil, false, err
		} else if field == nil {
			continue
		} else if field.EmbeddedSchema != nil {
			// If the field is an embedded schema, add the fields to the schema
			schema.Fields = append(schema.Fields, field.EmbeddedSchema.Fields...)
		} else {
			schema.Fields = append(schema.Fields, field)
		}
	}

	return schema, false, nil
}
