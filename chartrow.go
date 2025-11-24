package chatabase

import (
	"database/sql/driver"
	"time"

	"github.com/jmoiron/sqlx"
)

type ChartDataRow struct {
	XValue  interface{}   `json:"x_value"`
	YValues []interface{} `json:"y_values"`
}

// GetYValueAsFloat gets a Y-value as float64 with null safety
func (row *ChartDataRow) GetYValueAsFloat(index int) *float64 {
	if index >= len(row.YValues) || row.YValues[index] == nil {
		return nil
	}

	switch v := row.YValues[index].(type) {
	case float64:
		return &v
	case int64:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	default:
		return nil
	}
}

// ScanDynamicChart scans chart data with an unknown number of Y-values
func ScanDynamicChart(rows *sqlx.Rows) ([]ChartDataRow, error) {
	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []ChartDataRow

	for rows.Next() {
		// Create slice to hold all values (x + all y values)
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		// Create pointers to the values
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Convert byte arrays to appropriate types if needed
		for i, val := range values {
			values[i] = convertValue(val)
		}

		// Build the result row
		row := ChartDataRow{
			XValue:  values[0],  // First column is always x_value
			YValues: values[1:], // Rest are y_values
		}
		yValues := make([]interface{}, len(columns)-1)
		for i := range yValues {
			yValues[i] = map[string]interface{}{
				columns[i+1]: values[i+1],
			}
		}
		row.YValues = yValues
		results = append(results, row)
	}

	return results, rows.Err()
}

// convertValue converts database values to appropriate Go types
func convertValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	// Handle different database driver value types
	switch v := val.(type) {
	case []byte:
		// Convert byte slice to string
		return string(v)
	case driver.Value:
		return v
	case int64:
		return float64(v) // Convert integers to float64 for consistency
	case int32:
		return float64(v)
	case int:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		return v
	case time.Time:
		return v
	case bool:
		if v {
			return float64(1)
		}
		return float64(0)
	default:
		return v
	}
}
