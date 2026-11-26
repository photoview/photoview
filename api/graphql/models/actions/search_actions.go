package actions

import (
	"strings"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MaxSearchResults is the most a single search returns per category, whatever
// the caller asks for. A limit of 0 - "no limit" in the user's preference -
// means this many, as does any negative or larger value.
//
// The UI never asks for more than 500, so for the app this changes nothing.
// It is there for the API itself: the limit arguments are open to every
// authenticated client, and without a ceiling an empty query with a limit of 0
// would load and serialize a user's entire library in one request.
var MaxSearchResults = 1000

// boundedSearchLimit resolves a requested limit, nil falling back to fallback,
// to one between 1 and MaxSearchResults.
func boundedSearchLimit(requested *int, fallback int) int {
	limit := fallback
	if requested != nil {
		limit = *requested
	}

	if limit <= 0 || limit > MaxSearchResults {
		return MaxSearchResults
	}

	return limit
}

func Search(db *gorm.DB, query string, userID int, limitMedia *int, limitAlbums *int) (*models.SearchResult, error) {
	limitMediaInternal := boundedSearchLimit(limitMedia, 10)
	limitAlbumsInternal := boundedSearchLimit(limitAlbums, 10)

	wildQuery := "%" + strings.ToLower(query) + "%"

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

	mediaQuery = mediaQuery.Limit(limitMediaInternal)

	if err := mediaQuery.Find(&media).Error; err != nil {
		return nil, errors.Wrapf(err, "searching media")
	}

	var albums []*models.Album

	// LOWER on both sides, like the media query above: wildQuery is already
	// lowercased, and PostgreSQL's LIKE is case-sensitive, so without this an
	// album named "Summer" would never match a search for "summer".
	albumsQuery := db.
		Where("EXISTS (?)", db.Table("user_albums").Where("user_id = ?", userID).Where("album_id = albums.id")).
		Where("LOWER(albums.title) LIKE ? OR LOWER(albums.path) LIKE ?", wildQuery, wildQuery).
		Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:                "(CASE WHEN LOWER(albums.title) LIKE ? THEN 2 WHEN LOWER(albums.path) LIKE ? THEN 1 END) DESC, albums.title DESC, albums.id DESC",
				Vars:               []interface{}{wildQuery, wildQuery},
				WithoutParentheses: true},
		})

	albumsQuery = albumsQuery.Limit(limitAlbumsInternal)

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
