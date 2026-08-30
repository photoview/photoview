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

	if order != nil && order.OrderBy != nil {
		desc := false
		if order.OrderDirection != nil && order.OrderDirection.IsValid() {
			if *order.OrderDirection == OrderDirectionDesc {
				desc = true
			}
		}

		// Natural (numeric-aware) sorting for album titles.
		// Extracts leading numbers, pads them to 10 digits so "album_2" sorts before "album_10".
		// Falls back to alphabetical sort for titles that don't start with a number.
		if *order.OrderBy == "title_natural" {
			naturalExpr := `CASE WHEN title ~ '^[0-9]+' THEN LPAD(SUBSTRING(title FROM '^[0-9]+'), 10, '0') ELSE LOWER(title) END`
			if desc {
				tx.Order(clause.OrderByColumn{
					Column: clause.Column{Name: naturalExpr, Raw: true},
					Desc:   true,
				})
				tx.Order(clause.OrderByColumn{
					Column: clause.Column{Name: "LOWER(title)", Raw: true},
					Desc:   true,
				})
			} else {
				tx.Order(clause.OrderByColumn{
					Column: clause.Column{Name: naturalExpr, Raw: true},
				})
				tx.Order(clause.OrderByColumn{
					Column: clause.Column{Name: "LOWER(title)", Raw: true},
				})
			}
		} else {
			tx.Order(clause.OrderByColumn{
				Column: clause.Column{
					Name: *order.OrderBy,
				},
				Desc: desc,
			})
		}
	}

	return tx
}

// MD5Hash hashes value to a 32 length digest, the result is the same as the MYSQL function md5()
func MD5Hash(value string) string {
	hash := md5.Sum([]byte(value))
	return hex.EncodeToString(hash[:])
}
