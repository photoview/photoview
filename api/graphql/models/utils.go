package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func FormatSQL(tx *gorm.DB, order *Ordering, paginate *Pagination) *gorm.DB {

	if paginate != nil {
		if paginate.Limit != nil {
			tx.Limit(*paginate.Limit)
		}

		if paginate.Offset != nil {
			tx.Offset(*paginate.Offset)
		}
	}

	if *order.OrderBy == "random" {
		tx.Order("RAND()")
	} else {
		tx.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: *order.OrderBy,
			},
			Desc: desc,
		})
	}

	if order != nil && order.OrderBy != nil {
		desc := false
		if order.OrderDirection != nil && order.OrderDirection.IsValid() {
			if *order.OrderDirection == OrderDirectionDesc {
				desc = true
			}
		}

		tx.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: *order.OrderBy,
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
