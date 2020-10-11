package models

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

func (filter *Filter) FormatSQL() (string, error) {
	if filter == nil {
		return "", nil
	}

	result := ""

	if filter.OrderBy != nil {
		order_by := filter.OrderBy
		match, err := regexp.MatchString("^(\\w+(?:\\.\\w+)?(,\\s)?)+$", strings.TrimSpace(*filter.OrderBy))
		if err != nil {
			return "", err
		}

		if match {
			direction := "ASC"
			if filter.OrderDirection != nil && filter.OrderDirection.IsValid() {
				direction = filter.OrderDirection.String()
			}

			result += fmt.Sprintf(" ORDER BY %s %s", *order_by, direction)
		}
	}

	if filter.Limit != nil {
		offset := 0
		if filter.Offset != nil && *filter.Offset >= 0 {
			offset = *filter.Offset
		}

		result += fmt.Sprintf(" LIMIT %d OFFSET %d", *filter.Limit, offset)
	}

	log.Printf("SQL Filter: '%s'\n", result)
	return result, nil
}
