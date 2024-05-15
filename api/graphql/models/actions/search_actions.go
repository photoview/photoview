package actions

import (
	"strings"

	"github.com/kkovaletp/photoview/api/database/drivers"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Search(db *gorm.DB, query string, userID int, _limitMedia *int, _limitAlbums *int) (*models.SearchResult, error) {
	limitMedia := 10
	limitAlbums := 10

	if _limitMedia != nil {
		limitMedia = *_limitMedia
	}

	if _limitAlbums != nil {
		limitAlbums = *_limitAlbums
	}

	wildQuery := "%" + strings.ToLower(query) + "%"

	var media []*models.Media

	userSubquery := db.Table("user_albums").Where("user_id = ?", userID)
	if drivers.POSTGRES.MatchDatabase(db) {
		userSubquery = userSubquery.Where("album_id = \"Album\".id")
	} else {
		userSubquery = userSubquery.Where("album_id = Album.id")
	}

	err := db.Joins("Album").
		Where("EXISTS (?)", userSubquery).
		Where("LOWER(media.title) LIKE ? OR LOWER(media.path) LIKE ?", wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN LOWER(media.title) LIKE ? THEN 2 WHEN LOWER(media.path) LIKE ? THEN 1 END) DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		}).
		Limit(limitMedia).Find(&media).Error

	if err != nil {
		return nil, errors.Wrapf(err, "searching media")
	}

	var albums []*models.Album

	err = db.
		Where("EXISTS (?)", db.Table("user_albums").Where("user_id = ?", userID).Where("album_id = albums.id")).
		Where("albums.title LIKE ? OR albums.path LIKE ?", wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN albums.title LIKE ? THEN 2 WHEN albums.path LIKE ? THEN 1 END) DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		}).
		Limit(limitAlbums).
		Find(&albums).Error

	if err != nil {
		return nil, errors.Wrapf(err, "searching albums")
	}

	result := models.SearchResult{
		Query:  query,
		Media:  media,
		Albums: albums,
	}

	return &result, nil
}
