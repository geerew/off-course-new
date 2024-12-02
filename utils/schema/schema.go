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

	// A slice of relations
	Relations []*relation

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
	rv, err := concreteReflectValue(reflect.ValueOf(model))
	if err != nil {
		return nil, err
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

	for i := range rt.NumField() {
		sf := rt.Field(i)

		if fieldConfig, ok := config.fields[sf.Name]; ok {
			s.Fields = append(s.Fields, parseField(sf, fieldConfig))
		} else if _, ok := config.embedded[sf.Name]; ok {
			fields, err := parseEmbeddedField(sf)
			if err != nil {
				return nil, err
			}

			s.Fields = append(s.Fields, fields...)
		} else if relationConfig, ok := config.relations[sf.Name]; ok {
			s.Relations = append(s.Relations, parseRelation(sf, relationConfig))
		}
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
// CALLERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count calls the CountBuilder and executes the query, returning the count
func (s *Schema) Count(options *database.Options, db database.Querier) (int, error) {
	query, args, _ := s.CountBuilder(options).ToSql()

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Insert calls the InsertBuilder and executes the query, inserting a row
func (s *Schema) Insert(model any, db database.Querier) (sql.Result, error) {
	query, args, _ := s.InsertBuilder(model).ToSql()
	return db.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Select calls the SelectBuilder and executes the query, scanning the result into the model,
// which may be a struct or a slice of structs
func (s *Schema) Select(model any, options *database.Options, db database.Querier) error {
	rv := reflect.ValueOf(model)

	if rv.Kind() != reflect.Ptr {
		return utils.ErrNotPtr
	}

	if rv.IsNil() {
		return utils.ErrNilPtr
	}

	var err error
	var rows Rows

	concreteRv, err := concreteReflectValue(reflect.ValueOf(model))
	if err != nil {
		return err
	}

	if concreteRv.Kind() == reflect.Slice {
		query, args, _ := s.SelectBuilder(options).ToSql()

		rows, err = db.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		err = s.ScanMany(rows, rv, false)
		if err != nil {
			return err
		}

		err = s.loadRelationsMany(concreteRv, db)
	} else {
		query, args, _ := s.SelectBuilder(options).Limit(1).ToSql()
		rows, err = db.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		err = s.ScanOne(rows, rv, false)
		if err != nil {
			return err
		}

		err = s.loadRelationsOne(concreteRv, db)
	}

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Pluck calls the SelectBuilder and executes the query, scanning the result into the model,
// which may be a
func (s *Schema) Pluck(column string, result any, options *database.Options, db database.Querier) error {
	rv := reflect.ValueOf(result)

	if rv.Kind() != reflect.Ptr {
		return utils.ErrNotPtr
	}

	if rv.IsNil() {
		return utils.ErrNilPtr
	}

	var err error
	var rows Rows

	if column == "" || s.FieldsByColumn[column] == nil {
		return utils.ErrInvalidColumn
	}

	concreteRv, err := concreteReflectValue(reflect.ValueOf(result))
	if err != nil {
		return err
	}

	if concreteRv.Kind() == reflect.Slice {
		query, args, _ := s.PluckBuilder(column, options).ToSql()

		rows, err = db.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		err = s.ScanMany(rows, rv, true)
		if err != nil {
			return err
		}
	} else {
		query, args, _ := s.PluckBuilder(column, options).Limit(1).ToSql()
		rows, err = db.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		err = s.ScanOne(rows, rv, true)
		if err != nil {
			return err
		}
	}

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Insert calls the InsertBuilder and executes the query, inserting a row
func (s *Schema) Update(model any, options *database.Options, db database.Querier) (sql.Result, error) {
	builder, err := s.UpdateBuilder(model, options)
	if err != nil {
		return nil, err
	}

	query, args, _ := builder.ToSql()
	return db.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete calls the DeleteBuilder and executes the query, deleting rows
func (s *Schema) Delete(options *database.Options, db database.Querier) (sql.Result, error) {
	query, args, _ := s.DeleteBuilder(options).ToSql()
	return db.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// BUILDERS
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

// PluckBuilder creates a squirrel SelectBuilder for the model
func (s *Schema) PluckBuilder(column string, options *database.Options) squirrel.SelectBuilder {
	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(s.Table).
		RemoveColumns()

	for _, f := range s.Fields {
		if f.Column != column {
			continue
		}

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

		if options.Having != nil {
			builder = builder.Having(options.Having)
		}

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
func (s *Schema) UpdateBuilder(model any, options *database.Options) (squirrel.UpdateBuilder, error) {
	if options == nil || options.Where == nil {
		return squirrel.UpdateBuilder{}, utils.ErrInvalidWhere
	}

	builder := squirrel.
		StatementBuilder.
		Update(s.Table).
		Where(options.Where)

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

	return builder, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteBuilder creates a squirrel DeleteBuilder for the model
func (s *Schema) DeleteBuilder(options *database.Options) squirrel.DeleteBuilder {
	builder := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Delete(s.Table)

	if options != nil && options.Where != nil {
		builder = builder.Where(options.Where)
	}

	return builder
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// SCANS
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (s *Schema) ScanMany(rows Rows, rv reflect.Value, pluck bool) error {
	defer rows.Close()

	if rv.Kind() != reflect.Ptr {
		return utils.ErrNotPtr
	}

	if rv.IsNil() {
		return utils.ErrNilPtr
	}

	concreteRv, err := concreteReflectValue(rv)
	if err != nil {
		return err
	}

	if concreteRv.Kind() != reflect.Slice {
		return utils.ErrNotSlice
	}

	concreteRv.SetLen(0)

	isPtr := concreteRv.Type().Elem().Kind() == reflect.Ptr

	base := concreteRv.Type().Elem()
	if isPtr {
		base = base.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if pluck && len(columns) != 1 {
		return utils.ErrInvalidPluck
	}

	// Create a pointer of values for each field
	values := make([]interface{}, len(columns))

	for rows.Next() {
		instance := reflect.New(base)
		concreteInstance := reflect.Indirect(instance)

		if pluck {
			values[0] = concreteInstance.Addr().Interface()
		} else {
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
		}

		err = rows.Scan(values...)
		if err != nil {
			return err
		}

		if isPtr {
			concreteRv.Set(reflect.Append(concreteRv, instance))
		} else {
			concreteRv.Set(reflect.Append(concreteRv, concreteInstance))
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ScanOne scans a single row into the model
func (s *Schema) ScanOne(rows Rows, rv reflect.Value, pluck bool) error {
	defer rows.Close()

	if rv.Kind() != reflect.Ptr {
		return utils.ErrNotPtr
	}

	if rv.IsNil() {
		return utils.ErrNilPtr
	}

	concreteRv, err := concreteReflectValue(rv)
	if err != nil {
		return err
	}

	if concreteRv.Kind() != reflect.Struct {
		return utils.ErrNotStruct
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if pluck && len(columns) != 1 {
		return utils.ErrInvalidPluck
	}

	// Create a pointer of values for each field
	values := make([]interface{}, len(columns))

	if pluck {
		values[0] = concreteRv.Addr().Interface()
	} else {
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
	}

	if !rows.Next() {
		return sql.ErrNoRows
	}

	err = rows.Scan(values...)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// RELATIONS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loadRelationsOne loads the relations for a single model, handling both one-to-one and
// one-to-many relationships
func (s *Schema) loadRelationsOne(concreteRv reflect.Value, db database.Querier) error {
	if concreteRv.Kind() != reflect.Struct {
		return utils.ErrNotStruct
	}

	// Get the value of the ID field
	id, zero := s.FieldsByColumn["id"].ValueOf(concreteRv)
	if zero {
		return nil
	}

	for _, rel := range s.Relations {
		relatedSchema, relatedModelPtr, err := parseRelatedSchema(rel)
		if err != nil {
			return err
		}

		// Get the field in the struct to set the related model on
		structField := getStructField(concreteRv, rel.Position)

		// Create the options to select related rows
		options := &database.Options{Where: squirrel.Eq{rel.MatchOn: id}}

		if rel.HasMany {
			structFieldType := structField.Type()
			var structFieldPtr reflect.Value

			if rel.IsPtr {
				// Create a new pointer slice
				elemType := structFieldType.Elem()
				structField.Set(reflect.New(elemType))
				structField.Elem().Set(reflect.MakeSlice(elemType, 0, 0))
				structFieldPtr = structField.Elem().Addr()
			} else {
				// Create a new slice
				structField.Set(reflect.MakeSlice(structFieldType, 0, 0))
				structFieldPtr = structField.Addr()
			}

			err = relatedSchema.Select(structFieldPtr.Interface(), options, db)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
		} else {
			err = relatedSchema.Select(relatedModelPtr.Interface(), options, db)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				}

				return err
			}

			setRelatedField(structField, relatedModelPtr)
		}
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loadRelationsMany loads the relations for a slice of models, handling both many-to-one and
// many-to-many relationships
func (s *Schema) loadRelationsMany(concreteRv reflect.Value, db database.Querier) error {
	if concreteRv.Kind() != reflect.Slice {
		return utils.ErrNotSlice
	}

	// If the slice is empty, return
	if concreteRv.Len() == 0 {
		return nil
	}

	// Get the IDs of all the models so we only do 1 query per relation
	ids := []any{}
	for i := 0; i < concreteRv.Len(); i++ {
		v, zero := s.FieldsByColumn["id"].ValueOf(concreteRv.Index(i))
		if zero {
			continue
		}
		ids = append(ids, v)
	}

	for _, rel := range s.Relations {
		relatedSchema, _, err := parseRelatedSchema(rel)
		if err != nil {
			return err
		}

		// Create a slice to hold the related models
		relatedSlicePtr := reflect.New(reflect.SliceOf(rel.RelatedType))
		relatedSlice := relatedSlicePtr.Interface()

		// Create the options to select related rows
		options := &database.Options{Where: squirrel.Eq{rel.MatchOn: ids}}
		err = relatedSchema.Select(relatedSlice, options, db)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}

			return err
		}

		relatedSliceValue := reflect.Indirect(reflect.ValueOf(relatedSlice))

		if rel.HasMany {
			// --- MANY TO MANY ---

			// Create a map of where the key is the MatchOn value and the value is a slice of
			// related items
			resultMap := make(map[any][]reflect.Value)
			for i := 0; i < relatedSliceValue.Len(); i++ {
				item := relatedSliceValue.Index(i)
				id, _ := relatedSchema.FieldsByColumn[rel.MatchOn].ValueOf(reflect.Indirect(item))
				resultMap[id] = append(resultMap[id], item)
			}

			for i := 0; i < concreteRv.Len(); i++ {
				sliceItem := concreteRv.Index(i)

				v, zero := s.FieldsByColumn["id"].ValueOf(sliceItem)
				if zero {
					continue
				}

				relatedItems, found := resultMap[v]
				if !found {
					continue
				}

				// Get the field in the struct to set result
				structField := getStructField(sliceItem, rel.Position)
				structFieldType := structField.Type()

				// Create a new slice that will hold the related items, based upon whether it is a
				// slice of a pointer slice
				var newRelatedSlice reflect.Value
				if structFieldType.Kind() == reflect.Ptr {
					newRelatedSlice = reflect.MakeSlice(structFieldType.Elem(), 0, len(relatedItems))
				} else {
					newRelatedSlice = reflect.MakeSlice(structFieldType, 0, len(relatedItems))
				}

				// Append the related items to the new slice
				for _, item := range relatedItems {
					newRelatedSlice = reflect.Append(newRelatedSlice, item)
				}

				if structField.Kind() == reflect.Ptr {
					structField.Set(reflect.New(structFieldType.Elem()))
					structField.Elem().Set(newRelatedSlice)
				} else {
					structField.Set(newRelatedSlice)
				}
			}
		} else {
			// --- MANY TO ONE ---

			// Create a map of related items where the key is the MatchOn value
			relatedMap := make(map[any]reflect.Value)
			for i := 0; i < relatedSliceValue.Len(); i++ {
				item := relatedSliceValue.Index(i)
				id, _ := relatedSchema.FieldsByColumn[rel.MatchOn].ValueOf(reflect.Indirect(item))
				relatedMap[id] = item
			}

			// Set the related items on the model
			for i := 0; i < concreteRv.Len(); i++ {
				sliceItem := concreteRv.Index(i)

				v, zero := s.FieldsByColumn["id"].ValueOf(sliceItem)
				if zero {
					continue
				}

				relatedItem, found := relatedMap[v]
				if !found {
					continue
				}

				// Get the field in the struct and set the value
				relatedField := getStructField(sliceItem, rel.Position)
				setRelatedField(relatedField, relatedItem)
			}
		}
	}

	return nil
}
