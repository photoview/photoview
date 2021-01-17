package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (filter *Filter) FormatSQL(tx *gorm.DB) *gorm.DB {

	if filter == nil {
		return tx
	}

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

// MD5Hash hashes value to a 32 length digest, the result is the same as the MYSQL function md5()
func MD5Hash(value string) string {
	hash := md5.Sum([]byte(value))
	return hex.EncodeToString(hash[:])
}
