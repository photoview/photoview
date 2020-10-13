package models

import (
	"fmt"
	"log"
)

func (filter *Filter) FormatSQL(context string) (string, error) {
	if filter == nil {
		return "", nil
	}

	orderByMap := make(map[string]string)
	orderByMap["media_date_shot"] = "media.date_shot"
	orderByMap["media_date_imported"] = "media.date_imported"
	orderByMap["media_title"] = "media.title"
	orderByMap["media_kind"] = "media.media_type, SUBSTRING_INDEX(media.path, '.', -1)"
	orderByMap["album_title"] = "album.title"

	result := ""

	if filter.OrderBy != nil {
		order_by, ok := orderByMap[context+"_"+*filter.OrderBy]
		if !ok {
			log.Printf("Invalid order column: '%s'\n", *filter.OrderBy)
			return "", nil
		}

		direction := "ASC"
		if filter.OrderDirection != nil && filter.OrderDirection.IsValid() {
			direction = filter.OrderDirection.String()
		}

		result += fmt.Sprintf(" ORDER BY %s %s", order_by, direction)
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
