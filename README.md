# Chatabase

**Quick Query Generator for Charts**

Chatabase is a Go package that provides a powerful and flexible way to generate SQL queries for chart data visualization. It transforms declarative chart configurations into optimized SQL queries with proper NULL handling, boolean comparisons, and complex filtering capabilities.

## Features

- 🚀 **Quick Query Generation**: Transform chart configurations into SQL queries instantly
- 📊 **Multiple Chart Types**: Support for line, bar, pie, scatter, area, and histogram charts
- 🔄 **Complex Joins**: Handle multiple table joins with various join types
- 🎯 **Smart Filtering**: Advanced filtering with NULL handling and boolean comparisons
- 📈 **Aggregation Support**: Built-in aggregation functions (SUM, COUNT, AVG, MIN, MAX)
- 🛡️ **SQL Injection Safe**: Uses parameterized queries for security
- ✅ **Validation**: Comprehensive configuration validation
- 📁 **File I/O**: Load and save configurations from/to JSON files

## Installation

```bash
go get github.com/midedickson/chatabase
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/midedickson/chatabase"
)

func main() {
    // Create a chart configuration
    config := &chatabase.ChartConfig{
        ChartType: "line",
        Title:     "Monthly Sales",
        Tables: []chatabase.TableConfig{
            {
                Name: "orders",
                Joins: []chatabase.JoinConfig{
                    {
                        Table:     "users",
                        Type:      "LEFT",
                        Condition: "orders.user_id = users.id",
                    },
                },
            },
        },
        XAxis: chatabase.AxisConfig{
            Column:      "created_at",
            Aggregation: "DATE_TRUNC('month', created_at)",
            DataType:    "datetime",
        },
        YAxis: []chatabase.AxisConfig{
            {
                Column:      "amount",
                Aggregation: "SUM",
                DataType:    "numeric",
            },
        },
        GroupBy: []string{"DATE_TRUNC('month', created_at)"},
        Filters: []chatabase.FilterConfig{
            {
                Column:   "deleted_at",
                Operator: "IS NULL",
            },
            {
                Column:   "status",
                Operator: "=",
                Value:    "completed",
            },
        },
        OrderBy: []chatabase.OrderConfig{
            {
                Column:    "created_at",
                Direction: "ASC",
            },
        },
        Limit: 100,
    }

    // Generate SQL query
    query, args, err := chatabase.BuildChartQuery(*config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Query: %s\n", query)
    fmt.Printf("Args: %v\n", args)
}
```

## Configuration Structure

### ChartConfig

The main configuration structure that defines your chart:

```go
type ChartConfig struct {
    ChartType   string         `json:"chart_type"`   // "line", "bar", "pie", "scatter", "area", "histogram"
    Title       string         `json:"title"`
    Description string         `json:"description"`
    Tables      []TableConfig  `json:"tables"`
    XAxis       AxisConfig     `json:"x_axis"`
    YAxis       []AxisConfig   `json:"y_axis"`       // Multiple Y series support
    GroupBy     []string       `json:"group_by"`
    Filters     []FilterConfig `json:"filters"`
    Options     ChartOptions   `json:"options"`
    Limit       int            `json:"limit"`
    OrderBy     []OrderConfig  `json:"order_by"`
}
```

### Supported Chart Types

- **line**: Line charts for time series data
- **bar**: Bar charts for categorical comparisons
- **pie**: Pie charts for proportional data
- **scatter**: Scatter plots for correlation analysis
- **area**: Area charts for cumulative data
- **histogram**: Histograms for distribution analysis

### Table Configuration

```go
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
```

### Axis Configuration

```go
type AxisConfig struct {
    Column      string `json:"column"`           // Column name
    Label       string `json:"label"`            // Human-readable label
    Aggregation string `json:"aggregation"`      // "SUM", "COUNT", "AVG", "MIN", "MAX"
    DataType    string `json:"data_type"`        // "numeric", "datetime", "string"
    Format      string `json:"format,omitempty"` // "currency", "percentage", "date"
}
```

### Advanced Filtering

The package supports sophisticated filtering with automatic NULL and boolean handling:

```go
type FilterConfig struct {
    Column   string        `json:"column"`
    Operator string        `json:"operator"`
    Value    interface{}   `json:"value"`
    Values   []interface{} `json:"values,omitempty"` // For IN operator
}
```

#### Supported Operators

- **Equality**: `=`, `!=`, `<>`
- **Comparison**: `<`, `<=`, `>`, `>=`
- **Pattern Matching**: `LIKE`, `ILIKE`, `NOT LIKE`, `NOT ILIKE`
- **Set Operations**: `IN`, `NOT IN`
- **Range**: `BETWEEN`
- **NULL Checks**: `IS NULL`, `IS NOT NULL`

#### Smart NULL Handling

```go
// Automatically converts to "deleted_at IS NULL"
filter := FilterConfig{
    Column:   "deleted_at",
    Operator: "=",
    Value:    nil,
}

// Automatically converts to "is_active IS TRUE"
filter := FilterConfig{
    Column:   "is_active",
    Operator: "=",
    Value:    "true",
}
```

## Working with JSON

### Load Configuration from JSON

```go
// From JSON string
config, err := chatabase.UnmarshalChartConfig(jsonString)

// From file
config, err := chatabase.ParseChartConfigFromFile("chart_config.json")

// Multiple configurations
configs, err := chatabase.UnmarshalChartConfigs(jsonArrayString)
```

### Save Configuration to JSON

```go
// To JSON string
jsonString, err := chatabase.MarshalChartConfig(config)

// To file
err := chatabase.SaveChartConfigToFile(config, "chart_config.json")
```

### Example JSON Configuration

```json
{
  "chart_type": "line",
  "title": "Monthly Revenue Trend",
  "description": "Shows revenue trends over the past year",
  "tables": [
    {
      "name": "orders",
      "joins": [
        {
          "table": "users",
          "type": "LEFT",
          "condition": "orders.user_id = users.id"
        }
      ]
    }
  ],
  "x_axis": {
    "column": "created_at",
    "label": "Month",
    "aggregation": "DATE_TRUNC('month', created_at)",
    "data_type": "datetime"
  },
  "y_axis": [
    {
      "column": "amount",
      "label": "Revenue",
      "aggregation": "SUM",
      "data_type": "numeric",
      "format": "currency"
    }
  ],
  "group_by": ["DATE_TRUNC('month', created_at)"],
  "filters": [
    {
      "column": "deleted_at",
      "operator": "IS NULL"
    },
    {
      "column": "status",
      "operator": "IN",
      "values": ["completed", "shipped"]
    }
  ],
  "order_by": [
    {
      "column": "created_at",
      "direction": "ASC"
    }
  ],
  "limit": 12,
  "options": {
    "width": 800,
    "height": 400,
    "theme": "light",
    "show_legend": true,
    "show_grid": true
  }
}
```

## Validation

The package includes comprehensive validation:

```go
// Validate and normalize configuration
err := chatabase.ValidateAndNormalizeConfig(config)
if err != nil {
    // Handle validation errors
    fmt.Printf("Validation error: %v\n", err)
}
```

## Security Features

- **Parameterized Queries**: All user inputs are properly parameterized to prevent SQL injection
- **Input Validation**: Comprehensive validation of all configuration parameters
- **Type Safety**: Strong typing prevents common configuration errors

## Use Cases

- **Business Intelligence Dashboards**: Generate queries for KPI charts
- **Analytics Platforms**: Create flexible data visualization queries
- **Reporting Systems**: Build dynamic report queries
- **Data Exploration Tools**: Enable ad-hoc chart generation
- **Embedded Analytics**: Integrate chart generation into applications

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions, issues, or feature requests, please open an issue on GitHub.