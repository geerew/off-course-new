package schema

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Modeler defines the interface that each struct (model) should implement in order to be used
type Modeler interface {
	Table() string
	Definer
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Definer defines the interface that each struct (model) should implement in order to be used
type Definer interface {
	Define(*ModelConfig)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ModelConfig defines the configuration for the model
type ModelConfig struct {
	// Embedded fields
	embedded map[string]struct{}

	// Common fields
	fields map[string]*modelFieldConfig

	// Relations
	relations map[string]*modelRelationConfig

	// A slice of left joins
	leftJoins []*modelLeftJoinConfig
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Embedded field
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Embedded adds an embedded field
func (m *ModelConfig) Embedded(name string) {

	if m.embedded == nil {
		m.embedded = make(map[string]struct{})
	}

	m.embedded[name] = struct{}{}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Simple field
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// modelFieldConfig defines the configuration for a field in the model
type modelFieldConfig struct {
	// The name of the struct field
	name string
	// Override for the db column name
	column string
	// When true, the field cannot be null in the database
	notNull bool
	// A db column alias
	alias string
	// When true, the field can be updated
	mutable bool
	// When true, the field will be skipped during create if it is null
	ignoreIfNull bool
	// The table the field belongs to. When empty it belongs to the main table
	joinTable string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Field adds a common field to the fields map
func (m *ModelConfig) Field(name string) *modelFieldConfig {
	field := &modelFieldConfig{name: name}

	if m.fields == nil {
		m.fields = make(map[string]*modelFieldConfig)
	}

	m.fields[name] = field
	return field
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Column sets the name of the column in the database
func (f *modelFieldConfig) Column(name string) *modelFieldConfig {
	f.column = name
	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NotNull signals that the field cannot be null in the database
func (f *modelFieldConfig) NotNull() *modelFieldConfig {
	f.notNull = true
	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Alias sets the alias for the column
func (f *modelFieldConfig) Alias(name string) *modelFieldConfig {
	f.alias = name
	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Mutable signals that the field can be updated
func (f *modelFieldConfig) Mutable() *modelFieldConfig {
	f.mutable = true
	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IgnoreIfNull signals that the field should be skipped during create if it is null
func (f *modelFieldConfig) IgnoreIfNull() *modelFieldConfig {
	f.ignoreIfNull = true
	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// JoinTable sets the table the field belongs to
func (f *modelFieldConfig) JoinTable(table string) *modelFieldConfig {
	f.joinTable = table
	return f
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Relation field
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// modelFieldConfig defines the configuration for a field in the model
type modelRelationConfig struct {
	// The name of the struct field
	name string
	// The column on the relation table to match with
	match string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Field adds a relation field to the model
func (m *ModelConfig) Relation(name string) *modelRelationConfig {
	relation := &modelRelationConfig{name: name}

	if m.relations == nil {
		m.relations = make(map[string]*modelRelationConfig)
	}

	m.relations[name] = relation
	return relation
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Alias sets the alias for the column
func (r *modelRelationConfig) MatchOn(name string) *modelRelationConfig {
	r.match = name
	return r
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Left Join
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// modelLeftJoinConfig defines the configuration for a left join
type modelLeftJoinConfig struct {
	// The name of the table to join with
	table string
	// The condition for the join, e.g. "table1.id = table2.id"
	on string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// LeftJoin adds a left join to the model
func (m *ModelConfig) LeftJoin(table string) *modelLeftJoinConfig {
	join := &modelLeftJoinConfig{table: table}
	m.leftJoins = append(m.leftJoins, join)

	return join
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// On sets the condition for the join
func (j *modelLeftJoinConfig) On(condition string) *modelLeftJoinConfig {
	j.on = condition
	return j
}
