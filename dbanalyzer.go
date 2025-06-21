package chatabase

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

// ColumnInfo represents detailed information about a database column
type ColumnInfo struct {
	Name         string  `db:"column_name"`
	DataType     string  `db:"data_type"`
	IsNullable   string  `db:"is_nullable"`
	DefaultValue *string `db:"column_default"`
	MaxLength    *int    `db:"character_maximum_length"`
	Position     int     `db:"ordinal_position"`
	IsPrimaryKey bool    `db:"is_primary_key"`
	Comment      string  `db:"column_comment"`
}

// CustomType represents a PostgreSQL custom type
type CustomType struct {
	SchemaName  string  `db:"schema_name"`
	TypeName    string  `db:"type_name"`
	TypeType    string  `db:"type_type"`
	Category    string  `db:"category"`
	Owner       string  `db:"owner"`
	Description *string `db:"description"`
}

// EnumValue represents a value in an ENUM type
type EnumValue struct {
	TypeName  string `db:"type_name"`
	EnumLabel string `db:"enum_label"`
	SortOrder int    `db:"sort_order"`
}

// CompositeTypeAttribute represents an attribute of a composite type
type CompositeTypeAttribute struct {
	TypeName      string  `db:"type_name"`
	AttributeName string  `db:"attribute_name"`
	DataType      string  `db:"data_type"`
	Position      int     `db:"position"`
	IsNullable    bool    `db:"is_nullable"`
	DefaultValue  *string `db:"default_value"`
}

// DomainInfo represents a domain type with its constraints
type DomainInfo struct {
	SchemaName   string  `db:"schema_name"`
	DomainName   string  `db:"domain_name"`
	DataType     string  `db:"data_type"`
	IsNullable   bool    `db:"is_nullable"`
	DefaultValue *string `db:"default_value"`
	CheckClause  *string `db:"check_clause"`
	Description  *string `db:"description"`
}

type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

func GetTablesPostgreSQL(db *sqlx.DB) ([]string, error) {
	query := `
        SELECT tablename 
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY tablename`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

func GetColumnInfoPostgreSQL(db *sqlx.DB, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.ordinal_position,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			COALESCE(pgd.description, '') as column_comment
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku
				ON tc.constraint_name = ku.constraint_name
				AND tc.table_schema = ku.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = $1
				AND tc.table_schema = 'public'
		) pk ON c.column_name = pk.column_name
		LEFT JOIN pg_catalog.pg_statio_all_tables st 
			ON c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd 
			ON pgd.objoid = st.relid 
			AND pgd.objsubid = c.ordinal_position
		WHERE c.table_name = $1 
			AND c.table_schema = 'public'
		ORDER BY c.ordinal_position`

	var columns []ColumnInfo
	err := db.Select(&columns, query, tableName)
	return columns, err
}

// GetAllCustomTypes returns all custom types in the database
func GetAllCustomTypes(db *sqlx.DB, schemaName string) ([]CustomType, error) {
	query := `
		SELECT 
			n.nspname as schema_name,
			t.typname as type_name,
			CASE 
				WHEN t.typtype = 'c' THEN 'composite'
				WHEN t.typtype = 'e' THEN 'enum'
				WHEN t.typtype = 'd' THEN 'domain'
				WHEN t.typtype = 'b' THEN 'base'
				WHEN t.typtype = 'r' THEN 'range'
				WHEN t.typtype = 'p' THEN 'pseudo'
				ELSE 'unknown'
			END as type_type,
			CASE t.typcategory 
				WHEN 'A' THEN 'Array'
				WHEN 'B' THEN 'Boolean'
				WHEN 'C' THEN 'Composite'
				WHEN 'D' THEN 'Date/time'
				WHEN 'E' THEN 'Enum'
				WHEN 'G' THEN 'Geometric'
				WHEN 'I' THEN 'Network address'
				WHEN 'N' THEN 'Numeric'
				WHEN 'P' THEN 'Pseudo'
				WHEN 'R' THEN 'Range'
				WHEN 'S' THEN 'String'
				WHEN 'T' THEN 'Timespan'
				WHEN 'U' THEN 'User-defined'
				WHEN 'V' THEN 'Bit-string'
				WHEN 'X' THEN 'Unknown'
				ELSE 'Other'
			END as category,
			pg_get_userbyid(t.typowner) as owner,
			obj_description(t.oid, 'pg_type') as description
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE n.nspname = $1
			AND t.typtype IN ('c', 'e', 'd', 'b', 'r')  -- composite, enum, domain, base, range
			AND NOT EXISTS (
				SELECT 1 FROM pg_class c 
				WHERE c.reltype = t.oid AND c.relkind = 'c'
			)  -- Exclude table row types
		ORDER BY n.nspname, t.typname`

	var types []CustomType
	err := db.Select(&types, query, schemaName)
	return types, err
}

// GetEnumTypes returns all ENUM types with their values
func GetEnumTypes(db *sqlx.DB, schemaName string) (map[string][]EnumValue, error) {
	query := `
		SELECT 
			t.typname as type_name,
			e.enumlabel as enum_label,
			e.enumsortorder as sort_order
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		JOIN pg_enum e ON t.oid = e.enumtypid
		WHERE n.nspname = $1
			AND t.typtype = 'e'
		ORDER BY t.typname, e.enumsortorder`

	rows, err := db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	enumMap := make(map[string][]EnumValue)
	for rows.Next() {
		var ev EnumValue
		if err := rows.Scan(&ev.TypeName, &ev.EnumLabel, &ev.SortOrder); err != nil {
			return nil, err
		}
		enumMap[ev.TypeName] = append(enumMap[ev.TypeName], ev)
	}

	return enumMap, rows.Err()
}

// GetCompositeTypes returns all composite types with their attributes
func GetCompositeTypes(db *sqlx.DB, schemaName string) (map[string][]CompositeTypeAttribute, error) {
	query := `
		SELECT 
			t.typname as type_name,
			a.attname as attribute_name,
			format_type(a.atttypid, a.atttypmod) as data_type,
			a.attnum as position,
			NOT a.attnotnull as is_nullable,
			pg_get_expr(ad.adbin, ad.adrelid) as default_value
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		JOIN pg_class c ON c.reltype = t.oid
		JOIN pg_attribute a ON a.attrelid = c.oid
		LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
		WHERE n.nspname = $1
			AND t.typtype = 'c'
			AND a.attnum > 0
			AND NOT a.attisdropped
		ORDER BY t.typname, a.attnum`

	rows, err := db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	compositeMap := make(map[string][]CompositeTypeAttribute)
	for rows.Next() {
		var attr CompositeTypeAttribute
		if err := rows.Scan(&attr.TypeName, &attr.AttributeName, &attr.DataType,
			&attr.Position, &attr.IsNullable, &attr.DefaultValue); err != nil {
			return nil, err
		}
		compositeMap[attr.TypeName] = append(compositeMap[attr.TypeName], attr)
	}

	return compositeMap, rows.Err()
}

// GetDomainTypes returns all domain types with their constraints
func GetDomainTypes(db *sqlx.DB, schemaName string) ([]DomainInfo, error) {
	query := `
		SELECT 
			n.nspname as schema_name,
			t.typname as domain_name,
			format_type(t.typbasetype, t.typtypmod) as data_type,
			NOT t.typnotnull as is_nullable,
			t.typdefault as default_value,
			pg_get_constraintdef(cc.oid) as check_clause,
			obj_description(t.oid, 'pg_type') as description
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		LEFT JOIN pg_constraint cc ON cc.contypid = t.oid AND cc.contype = 'c'
		WHERE n.nspname = $1
			AND t.typtype = 'd'
		ORDER BY t.typname`

	var domains []DomainInfo
	err := db.Select(&domains, query, schemaName)
	return domains, err
}

// GetRangeTypes returns all range types
func GetRangeTypes(db *sqlx.DB, schemaName string) ([]CustomType, error) {
	query := `
		SELECT 
			n.nspname as schema_name,
			t.typname as type_name,
			'range' as type_type,
			'Range' as category,
			pg_get_userbyid(t.typowner) as owner,
			obj_description(t.oid, 'pg_type') as description
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE n.nspname = $1
			AND t.typtype = 'r'
		ORDER BY t.typname`

	var ranges []CustomType
	err := db.Select(&ranges, query, schemaName)
	return ranges, err
}

// GetCustomTypesWithDetails returns all custom types with their detailed information
func GetCustomTypesWithDetails(db *sqlx.DB, schemaName string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get all custom types
	allTypes, err := GetAllCustomTypes(db, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error getting custom types: %w", err)
	}
	result["all_types"] = allTypes

	// Get enum types with values
	enumTypes, err := GetEnumTypes(db, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error getting enum types: %w", err)
	}
	result["enum_types"] = enumTypes

	// Get composite types with attributes
	compositeTypes, err := GetCompositeTypes(db, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error getting composite types: %w", err)
	}
	result["composite_types"] = compositeTypes

	// Get domain types
	domainTypes, err := GetDomainTypes(db, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error getting domain types: %w", err)
	}
	result["domain_types"] = domainTypes

	// Get range types
	rangeTypes, err := GetRangeTypes(db, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error getting range types: %w", err)
	}
	result["range_types"] = rangeTypes

	return result, nil
}
