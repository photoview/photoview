package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (filter *Filter) FormatSQL(tx *gorm.DB) *gorm.DB {

	if filter.Limit != nil {
		tx.Limit(*filter.Limit)
	}

	if filter.Offset != nil {
		tx.Offset(*filter.Offset)
	}

	if filter.OrderBy != nil {

		desc := true
		if filter.OrderDirection != nil && filter.OrderDirection.IsValid() {
			if *filter.OrderDirection == OrderDirectionAsc {
				desc = false
			}
		}

		tx.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: *filter.OrderBy,
			},
			Desc: desc,
		})
	}

	return tx
}
