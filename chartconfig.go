package chatabase

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ChartConfig struct {
	// Chart basics
	ChartType   string `json:"chart_type"` // "line", "bar", "pie", "scatter", "area", "histogram"
	Title       string `json:"title"`
	Description string `json:"description"`

	// Data source
	Tables []TableConfig `json:"tables"`

	// Axes configuration
	XAxis AxisConfig   `json:"x_axis"`
	YAxis []AxisConfig `json:"y_axis"` // Array to support multiple Y series

	// Aggregation and grouping
	GroupBy []string       `json:"group_by"`
	Filters []FilterConfig `json:"filters"`

	// Chart-specific options
	Options ChartOptions `json:"options"`

	// Query limits
	Limit   int           `json:"limit"`
	OrderBy []OrderConfig `json:"order_by"`
}

type TableConfig struct {
	Name  string       `json:"name"`
	Alias string       `json:"alias,omitempty"`
	Joins []JoinConfig `json:"joins,omitempty"`
}

type JoinConfig struct {
	Table     string `json:"table"`
	Alias     string `json:"alias,omitempty"`
	Type      string `json:"type"`      // "INNER", "LEFT", "RIGHT", "FULL"
	Condition string `json:"condition"` // "users.id = orders.user_id"
}

type AxisConfig struct {
	Column      string `json:"column"`           // "created_at", "amount", "COUNT(*)"
	Label       string `json:"label"`            // Human-readable label
	Aggregation string `json:"aggregation"`      // "SUM", "COUNT", "AVG", "MIN", "MAX"
	DataType    string `json:"data_type"`        // "numeric", "datetime", "string"
	Format      string `json:"format,omitempty"` // "currency", "percentage", "date"
	Alias       string `json:"alias,omitempty"`  // NEW
}

type FilterConfig struct {
	Column    string        `json:"column"`
	Operator  string        `json:"operator"` // "=", "!=", ">", "<", ">=", "<=", "IN", "LIKE", "BETWEEN"
	Value     interface{}   `json:"value"`
	Values    []interface{} `json:"values,omitempty"` // For IN operator
	Raw       string        // NEW: if set, use as-is (with placeholders)
	RawValues []interface{} // NEW: bind params for Raw
}

type OrderConfig struct {
	Column    string `json:"column"`
	Direction string `json:"direction"` // "ASC", "DESC"
}

type ChartOptions struct {
	// Visual options
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Theme  string `json:"theme"`

	// Chart-specific
	Stacked    bool `json:"stacked,omitempty"` // For bar/area charts
	ShowLegend bool `json:"show_legend"`
	ShowGrid   bool `json:"show_grid"`

	// Date/time specific
	DateFormat   string `json:"date_format,omitempty"`
	TimeInterval string `json:"time_interval,omitempty"` // "day", "week", "month", "year"

	// Colors
	Colors []string `json:"colors,omitempty"`
}

// ParseMultipleConfigs parses multiple chart configurations from a JSON array string
func ParseMultipleConfigs(jsonArrayStr string) ([]*ChartConfig, error) {
	var rawConfigs []json.RawMessage

	if err := json.Unmarshal([]byte(jsonArrayStr), &rawConfigs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}

	configs := make([]*ChartConfig, 0, len(rawConfigs))

	for i, rawConfig := range rawConfigs {
		config, err := UnmarshalChartConfig(string(rawConfig))
		if err != nil {
			return nil, fmt.Errorf("failed to parse config at index %d: %w", i, err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// ValidateAndNormalizeConfig validates and normalizes a chart configuration
func ValidateAndNormalizeConfig(config *ChartConfig) error {
	// Validate first
	if err := validateChartConfig(config); err != nil {
		return err
	}

	// Normalize chart type to lowercase
	config.ChartType = strings.ToLower(config.ChartType)

	// Set default join type if not specified
	for i := range config.Tables {
		for j := range config.Tables[i].Joins {
			if config.Tables[i].Joins[j].Type == "" {
				config.Tables[i].Joins[j].Type = "INNER"
			}
		}
	}

	// Set default order direction if not specified
	for i := range config.OrderBy {
		if config.OrderBy[i].Direction == "" {
			config.OrderBy[i].Direction = "ASC"
		}
	}

	// Set default options
	if config.Options.Width == 0 {
		config.Options.Width = 800
	}
	if config.Options.Height == 0 {
		config.Options.Height = 400
	}
	if config.Options.Theme == "" {
		config.Options.Theme = "light"
	}

	return nil
}

// UnmarshalChartConfig unmarshals a JSON string into a ChartConfig struct
func UnmarshalChartConfig(jsonStr string) (*ChartConfig, error) {
	var config ChartConfig

	// Unmarshal the JSON string
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate the configuration
	if err := validateChartConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid chart configuration: %w", err)
	}

	return &config, nil
}

// MarshalChartConfig marshals a ChartConfig struct to a JSON string
func MarshalChartConfig(config *ChartConfig) (string, error) {
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ChartConfig: %w", err)
	}

	return string(jsonBytes), nil
}

// UnmarshalChartConfigs unmarshals a JSON array of chart configurations
func UnmarshalChartConfigs(jsonStr string) ([]*ChartConfig, error) {
	var configs []*ChartConfig

	if err := json.Unmarshal([]byte(jsonStr), &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON array: %w", err)
	}

	// Validate each configuration
	for i, config := range configs {
		if err := validateChartConfig(config); err != nil {
			return nil, fmt.Errorf("invalid chart configuration at index %d: %w", i, err)
		}
	}

	return configs, nil
}

// validateChartConfig validates the chart configuration
func validateChartConfig(config *ChartConfig) error {
	// Check required fields
	if config.ChartType == "" {
		return fmt.Errorf("chart_type is required")
	}

	if config.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(config.Tables) == 0 {
		return fmt.Errorf("at least one table is required")
	}

	if len(config.YAxis) == 0 {
		return fmt.Errorf("at least one y-axis is required")
	}

	// Validate chart type
	validChartTypes := []string{"line", "bar", "pie", "scatter", "area", "histogram"}
	if !contains(validChartTypes, config.ChartType) {
		return fmt.Errorf("invalid chart_type: %s. Must be one of: %s",
			config.ChartType, strings.Join(validChartTypes, ", "))
	}

	// Validate table configurations
	for i, table := range config.Tables {
		if table.Name == "" {
			return fmt.Errorf("table name is required at index %d", i)
		}

		// Validate joins
		for j, join := range table.Joins {
			if err := validateJoinConfig(&join, i, j); err != nil {
				return err
			}
		}
	}

	// Validate X-axis
	if config.XAxis.Column == "" {
		return fmt.Errorf("x_axis column is required")
	}

	// Validate Y-axes
	for i, yAxis := range config.YAxis {
		if yAxis.Column == "" {
			return fmt.Errorf("y_axis column is required at index %d", i)
		}

		if yAxis.Aggregation != "" {
			validAggregations := []string{"SUM", "COUNT", "AVG", "MIN", "MAX"}
			if !contains(validAggregations, yAxis.Aggregation) {
				return fmt.Errorf("invalid aggregation '%s' for y_axis at index %d. Must be one of: %s",
					yAxis.Aggregation, i, strings.Join(validAggregations, ", "))
			}
		}
	}

	// Validate filters
	for i, filter := range config.Filters {
		if err := validateFilter(&filter, i); err != nil {
			return err
		}
	}

	return nil
}

// validateJoinConfig validates a join configuration
func validateJoinConfig(join *JoinConfig, tableIndex, joinIndex int) error {
	if join.Table == "" {
		return fmt.Errorf("join table is required at table index %d, join index %d", tableIndex, joinIndex)
	}

	if join.Condition == "" {
		return fmt.Errorf("join condition is required at table index %d, join index %d", tableIndex, joinIndex)
	}

	validJoinTypes := []string{"INNER", "LEFT", "RIGHT", "FULL"}
	if join.Type != "" && !contains(validJoinTypes, join.Type) {
		return fmt.Errorf("invalid join type '%s' at table index %d, join index %d. Must be one of: %s",
			join.Type, tableIndex, joinIndex, strings.Join(validJoinTypes, ", "))
	}

	return nil
}

// validateFilter validates a filter configuration
func validateFilter(filter *FilterConfig, index int) error {
	if filter.Raw != "" {
		return nil
	}
	if filter.Column == "" {
		return fmt.Errorf("filter column is required at index %d", index)
	}

	if filter.Operator == "" {
		return fmt.Errorf("filter operator is required at index %d", index)
	}

	validOperators := []string{"=", "!=", ">", "<", ">=", "<=", "IN", "LIKE", "BETWEEN", "IS"}
	if !contains(validOperators, filter.Operator) {
		return fmt.Errorf("invalid filter operator '%s' at index %d. Must be one of: %s",
			filter.Operator, index, strings.Join(validOperators, ", "))
	}

	// Validate operator-specific requirements
	switch filter.Operator {
	case "IN":
		if len(filter.Values) == 0 {
			return fmt.Errorf("IN operator requires 'values' array at filter index %d", index)
		}
	case "BETWEEN":
		if len(filter.Values) != 2 {
			return fmt.Errorf("BETWEEN operator requires exactly 2 values at filter index %d", index)
		}
	case "IS":
		// IS typically used with NULL, allow both value and values to be empty
	default:
		if filter.Value == nil && len(filter.Values) == 0 {
			return fmt.Errorf("filter value is required for operator '%s' at index %d", filter.Operator, index)
		}
	}

	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ParseChartConfigFromFile reads and unmarshals a chart configuration from a file
func ParseChartConfigFromFile(filename string) (*ChartConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return UnmarshalChartConfig(string(data))
}

// SaveChartConfigToFile marshals and saves a chart configuration to a file
func SaveChartConfigToFile(config *ChartConfig, filename string) error {
	jsonStr, err := MarshalChartConfig(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, []byte(jsonStr), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}
