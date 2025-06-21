package chatabase

import (
	"fmt"
	"strings"
)

func BuildChartQuery(config *ChartConfig) (string, []interface{}, error) {
	var query strings.Builder
	var args []interface{}
	argIndex := 1

	// SELECT clause
	query.WriteString("SELECT ")

	// X-axis
	if config.XAxis.Aggregation != "" {
		query.WriteString(fmt.Sprintf("%s(%s) as x_value", config.XAxis.Aggregation, config.XAxis.Column))
	} else {
		query.WriteString(fmt.Sprintf("%s as x_value", config.XAxis.Column))
	}

	// Y-axis (multiple series support)
	for i, yAxis := range config.YAxis {
		query.WriteString(", ")
		if yAxis.Aggregation != "" {
			query.WriteString(fmt.Sprintf("%s(%s) as y_value_%d", yAxis.Aggregation, yAxis.Column, i))
		} else {
			query.WriteString(fmt.Sprintf("%s as y_value_%d", yAxis.Column, i))
		}
	}

	// FROM clause with joins
	query.WriteString(fmt.Sprintf(" FROM %s", config.Tables[0].Name))
	if config.Tables[0].Alias != "" {
		query.WriteString(fmt.Sprintf(" %s", config.Tables[0].Alias))
	}

	// JOINs
	for _, table := range config.Tables {
		for _, join := range table.Joins {
			query.WriteString(fmt.Sprintf(" %s JOIN %s", join.Type, join.Table))
			if join.Alias != "" {
				query.WriteString(fmt.Sprintf(" %s", join.Alias))
			}
			query.WriteString(fmt.Sprintf(" ON %s", join.Condition))
		}
	}

	// WHERE clause with NULL handling
	if len(config.Filters) > 0 {
		query.WriteString(" WHERE ")
		for i, filter := range config.Filters {
			if i > 0 {
				query.WriteString(" AND ")
			}

			switch strings.ToLower(filter.Operator) {
			case "in":
				// Handle NULL values in IN clause
				var nullCount int
				var nonNullValues []interface{}

				for _, val := range filter.Values {
					if val == nil {
						nullCount++
					} else {
						nonNullValues = append(nonNullValues, val)
					}
				}

				if nullCount > 0 && len(nonNullValues) > 0 {
					// Mix of NULL and non-NULL values
					query.WriteString("(")
					placeholders := make([]string, len(nonNullValues))
					for j, val := range nonNullValues {
						placeholders[j] = fmt.Sprintf("$%d", argIndex)
						args = append(args, val)
						argIndex++
					}
					query.WriteString(fmt.Sprintf("%s IN (%s) OR %s IS NULL",
						filter.Column, strings.Join(placeholders, ", "), filter.Column))
					query.WriteString(")")
				} else if nullCount > 0 {
					// Only NULL values
					query.WriteString(fmt.Sprintf("%s IS NULL", filter.Column))
				} else {
					// Only non-NULL values
					placeholders := make([]string, len(nonNullValues))
					for j, val := range nonNullValues {
						placeholders[j] = fmt.Sprintf("$%d", argIndex)
						args = append(args, val)
						argIndex++
					}
					query.WriteString(fmt.Sprintf("%s IN (%s)", filter.Column, strings.Join(placeholders, ", ")))
				}

			case "not in":
				// Handle NULL values in NOT IN clause
				var nullCount int
				var nonNullValues []interface{}

				for _, val := range filter.Values {
					if val == nil {
						nullCount++
					} else {
						nonNullValues = append(nonNullValues, val)
					}
				}

				if nullCount > 0 && len(nonNullValues) > 0 {
					// Mix of NULL and non-NULL values
					query.WriteString("(")
					placeholders := make([]string, len(nonNullValues))
					for j, val := range nonNullValues {
						placeholders[j] = fmt.Sprintf("$%d", argIndex)
						args = append(args, val)
						argIndex++
					}
					query.WriteString(fmt.Sprintf("%s NOT IN (%s) AND %s IS NOT NULL",
						filter.Column, strings.Join(placeholders, ", "), filter.Column))
					query.WriteString(")")
				} else if nullCount > 0 {
					// Only NULL values - NOT IN NULL means everything except NULL
					query.WriteString(fmt.Sprintf("%s IS NOT NULL", filter.Column))
				} else {
					// Only non-NULL values
					placeholders := make([]string, len(nonNullValues))
					for j, val := range nonNullValues {
						placeholders[j] = fmt.Sprintf("$%d", argIndex)
						args = append(args, val)
						argIndex++
					}
					query.WriteString(fmt.Sprintf("(%s NOT IN (%s) OR %s IS NULL)",
						filter.Column, strings.Join(placeholders, ", "), filter.Column))
				}

			case "between":
				// BETWEEN with NULL handling
				if filter.Values[0] == nil || filter.Values[1] == nil {
					return "", nil, fmt.Errorf("BETWEEN operator cannot have NULL values")
				}
				query.WriteString(fmt.Sprintf("%s BETWEEN $%d AND $%d", filter.Column, argIndex, argIndex+1))
				args = append(args, filter.Values[0], filter.Values[1])
				argIndex += 2

			case "=", "!=", "<>":
				// Handle NULL comparison
				if filter.Value == nil {
					if strings.ToLower(filter.Operator) == "=" {
						query.WriteString(fmt.Sprintf("%s IS NULL", filter.Column))
					} else {
						query.WriteString(fmt.Sprintf("%s IS NOT NULL", filter.Column))
					}
				} else {
					// Handle boolean comparison
					if boolVal, ok := filter.Value.(string); ok {
						lowerVal := strings.ToLower(boolVal)
						if lowerVal == "true" || lowerVal == "false" {
							if strings.ToLower(filter.Operator) == "=" {
								if lowerVal == "true" {
									query.WriteString(fmt.Sprintf("%s IS TRUE", filter.Column))
								} else {
									query.WriteString(fmt.Sprintf("%s IS FALSE", filter.Column))
								}
							} else { // != or <>
								if lowerVal == "true" {
									query.WriteString(fmt.Sprintf("%s IS NOT TRUE", filter.Column))
								} else {
									query.WriteString(fmt.Sprintf("%s IS NOT FALSE", filter.Column))
								}
							}
						} else {
							query.WriteString(fmt.Sprintf("%s %s $%d", filter.Column, filter.Operator, argIndex))
							args = append(args, filter.Value)
							argIndex++
						}
					} else if boolVal, ok := filter.Value.(bool); ok {
						// Handle actual boolean type
						if strings.ToLower(filter.Operator) == "=" {
							if boolVal {
								query.WriteString(fmt.Sprintf("%s IS TRUE", filter.Column))
							} else {
								query.WriteString(fmt.Sprintf("%s IS FALSE", filter.Column))
							}
						} else { // != or <>
							if boolVal {
								query.WriteString(fmt.Sprintf("%s IS NOT TRUE", filter.Column))
							} else {
								query.WriteString(fmt.Sprintf("%s IS NOT FALSE", filter.Column))
							}
						}
					} else {
						query.WriteString(fmt.Sprintf("%s %s $%d", filter.Column, filter.Operator, argIndex))
						args = append(args, filter.Value)
						argIndex++
					}
				}

			case "is null":
				query.WriteString(fmt.Sprintf("%s IS NULL", filter.Column))

			case "is not null":
				query.WriteString(fmt.Sprintf("%s IS NOT NULL", filter.Column))

			case "<", "<=", ">", ">=":
				// Comparison operators with NULL values
				if filter.Value == nil {
					return "", nil, fmt.Errorf("comparison operator %s cannot compare with NULL", filter.Operator)
				}
				query.WriteString(fmt.Sprintf("%s %s $%d", filter.Column, filter.Operator, argIndex))
				args = append(args, filter.Value)
				argIndex++

			case "like", "ilike", "not like", "not ilike":
				// LIKE operators with NULL handling
				if filter.Value == nil {
					query.WriteString(fmt.Sprintf("%s IS NULL", filter.Column))
				} else {
					query.WriteString(fmt.Sprintf("%s %s $%d", filter.Column, filter.Operator, argIndex))
					args = append(args, filter.Value)
					argIndex++
				}

			default:
				// Default case - handle NULL values
				if filter.Value == nil {
					query.WriteString(fmt.Sprintf("%s IS NULL", filter.Column))
				} else {
					query.WriteString(fmt.Sprintf("%s %s $%d", filter.Column, filter.Operator, argIndex))
					args = append(args, filter.Value)
					argIndex++
				}
			}
		}
	}

	// GROUP BY
	if len(config.GroupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(config.GroupBy, ", "))
	}

	// ORDER BY
	if len(config.OrderBy) > 0 {
		query.WriteString(" ORDER BY ")
		var orderClauses []string
		for _, order := range config.OrderBy {
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", order.Column, order.Direction))
		}
		query.WriteString(strings.Join(orderClauses, ", "))
	}

	// LIMIT
	if config.Limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", config.Limit))
	}

	return query.String(), args, nil
}
