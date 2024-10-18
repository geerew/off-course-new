package schema

import (
	"database/sql"
	"fmt"
	"reflect"
	"sync"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
)

var cache = &sync.Map{}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Schema defines the structure of a model after parsing
type Schema struct {
	// The name of the table in the database
	Table string

	// The fields of the model
	Fields []*field

	// FieldsByColumn is a map of fields by their DB column name
	FieldsByColumn map[string]*field

	// A slice of left joins
	LeftJoins []string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Parse parses a model
func Parse(model any) (*Schema, error) {
	if model == nil {
		return nil, utils.ErrNilPtr
	}

	// Get the reflect value and unwrap pointers
	rv := reflect.ValueOf(model)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() && rv.CanAddr() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}

		rv = rv.Elem()
	}

	// Error if the reflect value is invalid. This can happen if the model is nil, or an
	// uninitialized pointer like `var test *Test`
	if !rv.IsValid() {
		return nil, utils.ErrInvalidValue
	}

	rt := rv.Type()

	// If the model is a pointer, slice, or array, get the element type
	for rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array || rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	// Attempt to load the Schema from cache
	if v, ok := cache.Load(rt); ok {
		s := v.(*Schema)
		return s, nil
	}

	// Error when the model does not implement the Modeler interface
	modeler, isModeler := reflect.New(rt).Interface().(Modeler)
	if !isModeler {
		return nil, utils.ErrNotModeler
	}

	s := &Schema{
		Table: modeler.Table(),
	}

	config := &ModelConfig{}
	modeler.Define(config)

	if fields, err := parseFields(rt, config); err != nil {
		return nil, err
	} else {
		s.Fields = fields
	}

	// Build the FieldsByColumn map
	s.FieldsByColumn = make(map[string]*field, len(s.Fields))
	for _, f := range s.Fields {
		if f.Alias != "" {
			s.FieldsByColumn[f.Alias] = f
		} else {
			s.FieldsByColumn[f.Column] = f
		}
	}

	// Build the left joins
	for _, join := range config.leftJoins {
		s.LeftJoins = append(s.LeftJoins, fmt.Sprintf("%s ON %s", join.table, join.on))
	}

	// Store the schema in the cache
	if v, loaded := cache.LoadOrStore(rt, s); loaded {
		s := v.(*Schema)
		return s, nil
	}

	return s, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InsertBuilder creates a squirrel InsertBuilder for the model
func (s *Schema) InsertBuilder(model any) squirrel.InsertBuilder {
	data := make(map[string]any, len(s.Fields))

	for _, f := range s.Fields {
		// Ignore fields that part of a join
		if f.JoinTable != "" {
			continue
		}

		val, zero := f.ValueOf(reflect.ValueOf(model))

		// When the field cannot be null and the value is zero, set the value to nil
		if f.NotNull && zero {
			if f.IgnoreIfNull {
				continue
			}

			data[f.Column] = nil
		} else {
			data[f.Column] = val
		}
	}

	builder := squirrel.
		StatementBuilder.
		Insert(s.Table).
		SetMap(data)

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SelectBuilder creates a squirrel SelectBuilder for the model
func (s *Schema) SelectBuilder(options *database.Options) squirrel.SelectBuilder {
	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(s.Table).
		RemoveColumns()

	for _, f := range s.Fields {
		table := s.Table
		if f.JoinTable != "" {
			table = f.JoinTable
		}

		if f.Alias != "" {
			builder = builder.Column(fmt.Sprintf("%s.%s AS %s", table, f.Column, f.Alias))
		} else {
			builder = builder.Column(fmt.Sprintf("%s.%s", table, f.Column))
		}
	}

	for _, join := range s.LeftJoins {
		builder = builder.LeftJoin(join)
	}

	if options != nil {
		builder = builder.Where(options.Where).
			OrderBy(options.OrderBy...).
			GroupBy(options.GroupBy...)

		if options.Pagination != nil {
			builder = builder.
				Offset(uint64(options.Pagination.Offset())).
				Limit(uint64(options.Pagination.Limit()))
		}
	}

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateBuilder creates a squirrel UpdateBuilder for the model
func (s *Schema) UpdateBuilder(model any) squirrel.UpdateBuilder {
	builder := squirrel.
		StatementBuilder.
		Update(s.Table)

	for _, f := range s.Fields {
		if f.JoinTable != "" || !f.Mutable {
			continue
		}

		val, zero := f.ValueOf(reflect.ValueOf(model))

		if f.NotNull && zero {
			if f.IgnoreIfNull {
				continue
			}

			val = nil
		}

		builder = builder.Set(f.Column, val)
	}

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountBuilder creates a squirrel SelectBuilder for the model
func (s *Schema) CountBuilder(options *database.Options) squirrel.SelectBuilder {
	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("COUNT(DISTINCT " + s.Table + ".id)").
		From(s.Table)

	if options != nil && options.Where != nil {
		builder = builder.Where(options.Where)
	}

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteBuilder creates a squirrel DeleteBuilder for the model
func (s *Schema) DeleteBuilder(options *database.Options) squirrel.DeleteBuilder {
	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Delete(s.Table).
		Where(options.Where)

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Rows defines the interface for rows
type Rows interface {
	Columns() ([]string, error)
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
	Close() error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan scans the rows into the model, which can be a pointer to a slice or a single struct
func (s *Schema) Scan(rows Rows, model any) error {
	rv := reflect.ValueOf(model)
	defer rows.Close()

	if rv.Kind() != reflect.Ptr {
		return utils.ErrNotPtr
	}

	if rv.IsNil() {
		return utils.ErrNilPtr
	}

	concreteValue := reflect.Indirect(rv)

	if concreteValue.Kind() == reflect.Slice {
		concreteValue.SetLen(0)

		isPtr := concreteValue.Type().Elem().Kind() == reflect.Ptr

		base := concreteValue.Type().Elem()
		if isPtr {
			base = base.Elem()
		}

		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		// Create a pointer of values for each field
		values := make([]interface{}, len(columns))

		for rows.Next() {
			instance := reflect.New(base)
			concreteInstance := reflect.Indirect(instance)

			for idx, column := range columns {
				if field := s.FieldsByColumn[column]; field != nil {
					v := concreteInstance
					for _, pos := range field.Position {
						// TODO - If value is a pointer and nil, initialize it
						// TODO - If value is a map and nil, initialize it
						v = reflect.Indirect(v).Field(pos)
					}

					values[idx] = v.Addr().Interface()
				} else {
					return fmt.Errorf("column %s not found in model", column)
				}
			}

			err = rows.Scan(values...)
			if err != nil {
				return err
			}

			if isPtr {
				concreteValue.Set(reflect.Append(concreteValue, instance))
			} else {
				concreteValue.Set(reflect.Append(concreteValue, concreteInstance))
			}
		}

	} else {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		// Create a pointer of values for each field
		values := make([]interface{}, len(columns))

		for idx, column := range columns {
			if field := s.FieldsByColumn[column]; field != nil {
				v := rv
				for _, pos := range field.Position {
					// TODO - If value is a pointer and nil, initialize it
					// TODO - If value is a map and nil, initialize it
					v = reflect.Indirect(v).Field(pos)
				}

				values[idx] = v.Addr().Interface()
			} else {
				return fmt.Errorf("column %s not found in model", column)
			}
		}

		if !rows.Next() {
			return sql.ErrNoRows
		}

		err = rows.Scan(values...)
		if err != nil {
			return err
		}
	}

	return nil
}
