package actions

import (
	"strings"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Search(db *gorm.DB, query string, userID int, limitMedia *int, limitAlbums *int, showHidden *bool) (*models.SearchResult, error) {
	limitMediaInternal := 10
	limitAlbumsInternal := 10

	if limitMedia != nil {
		limitMediaInternal = *limitMedia
	}

	if limitAlbums != nil {
		limitAlbumsInternal = *limitAlbums
	}

	wildQuery := "%" + strings.ToLower(query) + "%"

	// "Hidden means fully hidden everywhere": a personally-hidden album (and
	// everything inside it) is excluded from search results too, unless the
	// viewer has asked to see hidden albums.
	var hiddenClosureIDs []int
	if showHidden == nil || !*showHidden {
		var err error
		hiddenClosureIDs, err = models.HiddenAlbumsClosure(db, userID)
		if err != nil {
			return nil, errors.Wrap(err, "computing hidden albums closure for search")
		}
	}

	var media []*models.Media

	userSubquery := db.Table("user_albums").Where("user_id = ?", userID)
	if drivers.POSTGRES.MatchDatabase(db) {
		userSubquery = userSubquery.Where("album_id = \"Album\".id")
	} else {
		userSubquery = userSubquery.Where("album_id = Album.id")
	}

	mediaQuery := db.Joins("Album").
		Where("EXISTS (?)", userSubquery).
		Where("LOWER(media.title) LIKE ? OR LOWER(media.path) LIKE ?", wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN LOWER(media.title) LIKE ? THEN 2 WHEN LOWER(media.path) LIKE ? THEN 1 END) DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		})

	if len(hiddenClosureIDs) > 0 {
		mediaQuery = mediaQuery.Where("media.album_id NOT IN (?)", hiddenClosureIDs)
	}

	// A limit of 0 or less means no limit, i.e. all matching results are returned.
	if limitMediaInternal > 0 {
		mediaQuery = mediaQuery.Limit(limitMediaInternal)
	}

	if err := mediaQuery.Find(&media).Error; err != nil {
		return nil, errors.Wrapf(err, "searching media")
	}

	var albums []*models.Album

	albumsQuery := db.
		Where("EXISTS (?)", db.Table("user_albums").Where("user_id = ?", userID).Where("album_id = albums.id")).
		Where("albums.title LIKE ? OR albums.path LIKE ?", wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN albums.title LIKE ? THEN 2 WHEN albums.path LIKE ? THEN 1 END) DESC, albums.title DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		})

	if len(hiddenClosureIDs) > 0 {
		albumsQuery = albumsQuery.Where("albums.id NOT IN (?)", hiddenClosureIDs)
	}

	if limitAlbumsInternal > 0 {
		albumsQuery = albumsQuery.Limit(limitAlbumsInternal)
	}

	if err := albumsQuery.Find(&albums).Error; err != nil {
		return nil, errors.Wrapf(err, "searching albums")
	}

	result := models.SearchResult{
		Query:  query,
		Media:  media,
		Albums: albums,
	}

	return &result, nil
}
