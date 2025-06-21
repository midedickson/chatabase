package chatabase

func ToSql(c *ChartConfig) (string, []interface{}, error) {
	err := ValidateAndNormalizeConfig(c)
	if err != nil {
		return "", nil, err
	}

	query, args, err := BuildChartQuery(c)
	if err != nil {
		return "", nil, err
	}

	return query, args, err
}
