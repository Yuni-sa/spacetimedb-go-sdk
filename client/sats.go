package client

import (
	"encoding/json"
	"fmt"
)

const SatsProtocol = "v1.json.spacetimedb"

// AlgebraicValue represents any SATS value
type AlgebraicValue interface {
	isAlgebraicValue()
}

// AlgebraicType represents any SATS type
type AlgebraicType struct {
	Sum     *SumType          `json:"Sum,omitempty"`
	Product *ProductType      `json:"Product,omitempty"`
	Builtin *BuiltinType      `json:"Builtin,omitempty"`
	Ref     *AlgebraicTypeRef `json:"Ref,omitempty"`
}

// Values

// SumValue represents an instance of a SumType (tagged union)
type SumValue struct {
	Tag   string         `json:"-"` // The variant tag
	Value AlgebraicValue `json:"-"` // The variant data
}

// MarshalJSON implements custom JSON marshaling for SumValue
func (sv SumValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]AlgebraicValue{
		sv.Tag: sv.Value,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for SumValue
func (sv *SumValue) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if len(m) != 1 {
		return fmt.Errorf("SumValue must have exactly one key-value pair")
	}

	for tag, rawValue := range m {
		sv.Tag = tag
		// Note: You'd need to unmarshal to specific type based on schema
		var value any
		if err := json.Unmarshal(rawValue, &value); err != nil {
			return err
		}
		sv.Value = BuiltinValue{Value: value}
		break
	}

	return nil
}

func (SumValue) isAlgebraicValue() {}

// ProductValue represents an instance of a ProductType (struct/tuple)
type ProductValue struct {
	Elements []AlgebraicValue `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for ProductValue
func (pv ProductValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(pv.Elements)
}

// UnmarshalJSON implements custom JSON unmarshaling for ProductValue
func (pv *ProductValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &pv.Elements)
}

func (ProductValue) isAlgebraicValue() {}

// BuiltinValue represents an instance of a BuiltinType (primitive type)
type BuiltinValue struct {
	Value any `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for BuiltinValue
func (bv BuiltinValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(bv.Value)
}

// UnmarshalJSON implements custom JSON unmarshaling for BuiltinValue
func (bv *BuiltinValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &bv.Value)
}

func (BuiltinValue) isAlgebraicValue() {}

// Types

// SumType represents sum types (tagged unions)
type SumType struct {
	Variants []SumTypeVariant `json:"variants"`
}

// SumTypeVariant represents a variant in a sum type
type SumTypeVariant struct {
	AlgebraicType AlgebraicType   `json:"algebraic_type"`
	Name          *OptionalString `json:"name,omitempty"`
}

// ProductType represents product types (structs/tuples)
type ProductType struct {
	Elements []ProductTypeElement `json:"elements"`
}

// ProductTypeElement represents an element in a product type
type ProductTypeElement struct {
	AlgebraicType AlgebraicType   `json:"algebraic_type"`
	Name          *OptionalString `json:"name,omitempty"`
}

// BuiltinType represents primitive SATS types
type BuiltinType struct {
	Bool   bool          `json:"Bool,omitempty"`
	I8     int8          `json:"I8,omitempty"`
	U8     uint8         `json:"U8,omitempty"`
	I16    int16         `json:"I16,omitempty"`
	U16    uint16        `json:"U16,omitempty"`
	I32    int32         `json:"I32,omitempty"`
	U32    uint32        `json:"U32,omitempty"`
	I64    int64         `json:"I64,omitempty"`
	U64    uint64        `json:"U64,omitempty"`
	I128   *struct{}     `json:"I128,omitempty"` // TODO: Implement I128
	U128   *struct{}     `json:"U128,omitempty"` // TODO: Implement U128
	F32    float32       `json:"F32,omitempty"`
	F64    float64       `json:"F64,omitempty"`
	String string        `json:"String,omitempty"`
	Array  AlgebraicType `json:"Array,omitempty"`
	Map    *MapType      `json:"Map,omitempty"`
}

// MapType represents a map type with key and value types
type MapType struct {
	KeyType   AlgebraicType `json:"key_ty"`
	ValueType AlgebraicType `json:"ty"`
}

// AlgebraicTypeRef represents an indirect reference to a type
type AlgebraicTypeRef int

// OptionalString represents an optional string value
type OptionalString struct {
	Some *string   `json:"some,omitempty"`
	None *struct{} `json:"none,omitempty"`
}

// NewSomeString creates an OptionalString with a value
func NewSomeString(value string) OptionalString {
	return OptionalString{Some: &value}
}

// NewNoneString creates an OptionalString without a value
func NewNoneString() OptionalString {
	return OptionalString{None: &struct{}{}}
}

// IsNone returns true if the OptionalString has no value
func (os OptionalString) IsNone() bool {
	return os.None != nil
}

// IsSome returns true if the OptionalString has a value
func (os OptionalString) IsSome() bool {
	return os.Some != nil
}

// Value returns the string value if present, empty string otherwise
func (os OptionalString) Value() string {
	if os.Some != nil {
		return *os.Some
	}
	return ""
}

// Builtin type constructors for convenience

// NewBoolType creates a boolean builtin type
func NewBoolType() BuiltinType {
	var b bool
	return BuiltinType{Bool: b}
}

// NewStringType creates a string builtin type
func NewStringType() BuiltinType {
	var s string
	return BuiltinType{String: s}
}

// NewI32Type creates an i32 builtin type
func NewI32Type() BuiltinType {
	var i32 int32
	return BuiltinType{I32: i32}
}

// NewU64Type creates a u64 builtin type
func NewU64Type() BuiltinType {
	var u64 uint64
	return BuiltinType{U64: u64}
}

// NewArrayType creates an array builtin type
func NewArrayType(elementType AlgebraicType) BuiltinType {
	return BuiltinType{Array: elementType}
}

// NewMapType creates a map builtin type
func NewMapType(keyType, valueType AlgebraicType) BuiltinType {
	return BuiltinType{Map: &MapType{
		KeyType:   keyType,
		ValueType: valueType,
	}}
}

// AlgebraicType constructors

// NewSumAlgebraicType creates an AlgebraicType from a SumType
func NewSumAlgebraicType(sumType SumType) AlgebraicType {
	return AlgebraicType{Sum: &sumType}
}

// NewProductAlgebraicType creates an AlgebraicType from a ProductType
func NewProductAlgebraicType(productType ProductType) AlgebraicType {
	return AlgebraicType{Product: &productType}
}

// NewBuiltinAlgebraicType creates an AlgebraicType from a BuiltinType
func NewBuiltinAlgebraicType(builtinType BuiltinType) AlgebraicType {
	return AlgebraicType{Builtin: &builtinType}
}

// NewRefAlgebraicType creates an AlgebraicType from a type reference
func NewRefAlgebraicType(ref AlgebraicTypeRef) AlgebraicType {
	return AlgebraicType{Ref: &ref}
}

// Type checking methods

// IsSum returns true if this is a sum type
func (at AlgebraicType) IsSum() bool {
	return at.Sum != nil
}

// IsProduct returns true if this is a product type
func (at AlgebraicType) IsProduct() bool {
	return at.Product != nil
}

// IsBuiltin returns true if this is a builtin type
func (at AlgebraicType) IsBuiltin() bool {
	return at.Builtin != nil
}

// IsRef returns true if this is a type reference
func (at AlgebraicType) IsRef() bool {
	return at.Ref != nil
}

// GetSum returns the SumType if this is a sum type
func (at AlgebraicType) GetSum() *SumType {
	return at.Sum
}

// GetProduct returns the ProductType if this is a product type
func (at AlgebraicType) GetProduct() *ProductType {
	return at.Product
}

// GetBuiltin returns the BuiltinType if this is a builtin type
func (at AlgebraicType) GetBuiltin() *BuiltinType {
	return at.Builtin
}

// GetRef returns the AlgebraicTypeRef if this is a type reference
func (at AlgebraicType) GetRef() *AlgebraicTypeRef {
	return at.Ref
}

// Typespace represents a collection of types with references
type Typespace struct {
	Types []AlgebraicType `json:"types"`
}

// GetType returns the type at the given index
func (ts Typespace) GetType(ref AlgebraicTypeRef) *AlgebraicType {
	index := int(ref)
	if index < 0 || index >= len(ts.Types) {
		return nil
	}
	return &ts.Types[index]
}

// AddType adds a type to the typespace and returns its reference
func (ts *Typespace) AddType(typ AlgebraicType) AlgebraicTypeRef {
	ts.Types = append(ts.Types, typ)
	return AlgebraicTypeRef(len(ts.Types) - 1)
}

// Schema-related types

// RawModuleDef represents a complete module definition
type RawModuleDef struct {
	Typespace        Typespace      `json:"typespace"`
	Tables           []TableDef     `json:"tables"`
	Reducers         []ReducerDef   `json:"reducers"`
	Types            []NamedTypeDef `json:"types"`
	MiscExports      []any          `json:"misc_exports"`
	RowLevelSecurity []any          `json:"row_level_security"`
}

// TableDef represents a table definition
type TableDef struct {
	Name           string           `json:"name"`
	ProductTypeRef AlgebraicTypeRef `json:"product_type_ref"`
	PrimaryKey     []any            `json:"primary_key"`
	Indexes        []any            `json:"indexes"`
	Constraints    []any            `json:"constraints"`
	Sequences      []any            `json:"sequences"`
	Schedule       ScheduleType     `json:"schedule"`
	TableType      TableType        `json:"table_type"`
	TableAccess    TableAccessType  `json:"table_access"`
}

// ScheduleType represents table scheduling options
type ScheduleType struct {
	None []any `json:"none"`
}

// TableType represents the type of table
type TableType struct {
	User   []any `json:"User,omitempty"`
	System []any `json:"System,omitempty"`
}

// TableAccessType represents table access permissions
type TableAccessType struct {
	Private []any `json:"Private,omitempty"`
	Public  []any `json:"Public,omitempty"`
}

// ReducerDef represents a reducer definition
type ReducerDef struct {
	Name      string           `json:"name"`
	Params    ProductType      `json:"params"`
	Lifecycle ReducerLifecycle `json:"lifecycle"`
}

// ReducerLifecycle represents when a reducer should be called
type ReducerLifecycle struct {
	None []any                  `json:"none,omitempty"`
	Some *ReducerLifecycleEvent `json:"some,omitempty"`
}

// ReducerLifecycleEvent represents specific lifecycle events
type ReducerLifecycleEvent struct {
	OnConnect    []any `json:"OnConnect,omitempty"`
	OnDisconnect []any `json:"OnDisconnect,omitempty"`
	Init         []any `json:"Init,omitempty"`
}

// NamedTypeDef represents a named type definition
type NamedTypeDef struct {
	Name           TypeName         `json:"name"`
	Type           AlgebraicTypeRef `json:"ty"`
	CustomOrdering bool             `json:"custom_ordering"`
}

// TypeName represents a scoped type name
type TypeName struct {
	Scope []string `json:"scope"`
	Name  string   `json:"name"`
}

// Convenience constructors for common types

// NewUserTable creates a user table definition
func NewUserTable(name string, typeRef AlgebraicTypeRef) TableDef {
	return TableDef{
		Name:           name,
		ProductTypeRef: typeRef,
		PrimaryKey:     []any{},
		Indexes:        []any{},
		Constraints:    []any{},
		Sequences:      []any{},
		Schedule:       ScheduleType{None: []any{}},
		TableType:      TableType{User: []any{}},
		TableAccess:    TableAccessType{Public: []any{}},
	}
}

// NewReducer creates a reducer definition
func NewReducer(name string, params ProductType) ReducerDef {
	return ReducerDef{
		Name:      name,
		Params:    params,
		Lifecycle: ReducerLifecycle{None: []any{}},
	}
}

// NewInitReducer creates an init lifecycle reducer
func NewInitReducer(name string, params ProductType) ReducerDef {
	return ReducerDef{
		Name:   name,
		Params: params,
		Lifecycle: ReducerLifecycle{
			Some: &ReducerLifecycleEvent{Init: []any{}},
		},
	}
}
