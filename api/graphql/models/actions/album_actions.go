package actions

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func MyAlbums(db *gorm.DB, user *models.User, order *models.Ordering, paginate *models.Pagination,
	onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool, showHidden *bool) ([]*models.Album, error) {

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	if len(user.Albums) == 0 {
		return nil, nil
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := db.Model(models.Album{}).Where("id IN (?)", userAlbumIDs)

	if onlyRoot != nil && *onlyRoot {

		singleRootAlbumID := getSingleRootAlbumID(user)

		if singleRootAlbumID != -1 && len(user.Albums) > 1 {
			query = query.Where("parent_album_id = ?", singleRootAlbumID)
		} else {
			query = query.Where("parent_album_id IS NULL OR parent_album_id NOT IN (?)", userAlbumIDs)
		}
	}

	query = favoritesQuery(showEmpty, db, onlyWithFavorites, user, query)
	query, err := HiddenAlbumsFilter(showHidden, db, user, query)
	if err != nil {
		return nil, err
	}

	query = models.FormatSQL(query, order, paginate)

	var albums []*models.Album
	if err := query.Find(&albums).Error; err != nil {
		return nil, err
	}

	return albums, nil
}

// getSingleRootAlbumID returns the ID of the user's only true root album
// (ParentAlbumID == nil) if it accounts for everything the user can see -
// i.e. every other album they have access to is a descendant of it. In
// that case, MyAlbums flattens the redundant root away and returns its
// children directly, sparing a click into an otherwise pointless single
// top-level folder.
//
// If the user has any other album whose own parent isn't in their album
// set either, that's an independent share unrelated to the single root
// (e.g. a folder shared from a different user's library, whose real
// parent the recipient has no access to) - this returns -1 so it shows
// up as its own top-level entry instead of silently disappearing under
// the single-root special case.
func getSingleRootAlbumID(user *models.User) int {
	var singleRootAlbumID int = -1
	for _, album := range user.Albums {
		if album.ParentAlbumID == nil {
			if singleRootAlbumID != -1 {
				return -1
			}
			singleRootAlbumID = album.ID
		}
	}
	if singleRootAlbumID == -1 {
		return -1
	}

	albumIDs := make(map[int]bool, len(user.Albums))
	for _, album := range user.Albums {
		albumIDs[album.ID] = true
	}

	for _, album := range user.Albums {
		if album.ID == singleRootAlbumID {
			continue
		}
		if album.ParentAlbumID == nil || !albumIDs[*album.ParentAlbumID] {
			return -1
		}
	}

	return singleRootAlbumID
}

// hiddenAlbumsFilter excludes albums the user has personally hidden, plus
// all descendants of a hidden album (they're unreachable by browsing once
// their ancestor is hidden, even though they have no hidden row of their
// own), unless showHidden is true (used to reveal hidden albums, dimmed, in
// the UI).
func HiddenAlbumsFilter(showHidden *bool, db *gorm.DB, user *models.User, query *gorm.DB) (*gorm.DB, error) {
	if showHidden != nil && *showHidden {
		return query, nil
	}

	hiddenClosureIDs, err := models.HiddenAlbumsClosure(db, user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "computing hidden albums closure")
	}

	if len(hiddenClosureIDs) > 0 {
		query = query.Where("id NOT IN (?)", hiddenClosureIDs)
	}

	return query, nil
}

func favoritesQuery(showEmpty *bool, db *gorm.DB, onlyWithFavorites *bool, user *models.User, query *gorm.DB) *gorm.DB {
	if showEmpty == nil || !*showEmpty {
		subQuery := db.Model(&models.Media{}).Where("album_id = albums.id")

		if onlyWithFavorites != nil && *onlyWithFavorites {
			favoritesSubquery := db.
				Model(&models.UserMediaData{UserID: user.ID}).
				Where("user_media_data.media_id = media.id").
				Where("user_media_data.favorite = true")

			subQuery = subQuery.Where("EXISTS (?)", favoritesSubquery)
		}

		query = query.Where("EXISTS (?)", subQuery)
	}
	return query
}

func Album(db *gorm.DB, user *models.User, id int) (*models.Album, error) {
	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("album not found")
		}
		return nil, err
	}

	hasAccess, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelRead)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("forbidden")
	}

	return &album, nil
}

func AlbumPath(db *gorm.DB, user *models.User, album *models.Album) ([]*models.Album, error) {
	var albumPath []*models.Album

	// depth is carried through the recursion and used only to ORDER BY -
	// SQL doesn't guarantee a recursive CTE returns rows in any particular
	// order otherwise, and the truncation loop below depends on seeing the
	// closest ancestor first.
	if err := db.Raw(`
		WITH recursive path_albums AS (
			SELECT *, 0 AS depth FROM albums anchor WHERE anchor.id = ?
			UNION
			SELECT parent.*, child.depth + 1 FROM path_albums child JOIN albums parent ON parent.id = child.parent_album_id
		)
		SELECT * FROM path_albums WHERE id != ? ORDER BY depth ASC
	`, album.ID, album.ID).Scan(&albumPath).Error; err != nil {
		return nil, err
	}

	// albumPath is ordered closest-ancestor-first, root-last. Access only
	// ever cascades downward (from a grant point to its descendants), so
	// walk outward from the leaf and stop at the first ancestor the user
	// can't see - everything from there to the root is truncated, while
	// closer, still-visible ancestors (e.g. a shared subfolder sitting
	// below an otherwise inaccessible root) are kept.
	visibleUpTo := len(albumPath)
	for i, ancestor := range albumPath {
		hasAccess, err := user.HasAlbumLevel(db, ancestor, models.AlbumPermissionLevelRead)
		if err != nil {
			return nil, err
		}

		if !hasAccess {
			visibleUpTo = i
			break
		}
	}

	return albumPath[:visibleUpTo], nil
}

func SetAlbumCover(db *gorm.DB, user *models.User, mediaID int) (*models.Album, error) {
	var media models.Media

	if err := db.Find(&media, mediaID).Error; err != nil {
		return nil, err
	}

	var album models.Album

	if err := db.Find(&album, &media.AlbumID).Error; err != nil {
		return nil, err
	}

	hasAccess, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelUpload)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("forbidden")
	}

	if err := db.Model(&album).Update("cover_id", mediaID).Error; err != nil {
		return nil, err
	}

	return &album, nil
}

func ResetAlbumCover(db *gorm.DB, user *models.User, albumID int) (*models.Album, error) {
	var album models.Album
	if err := db.Find(&album, albumID).Error; err != nil {
		return nil, err
	}

	hasAccess, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelUpload)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, errors.New("forbidden")
	}

	if err := db.Model(&album).Update("cover_id", nil).Error; err != nil {
		return nil, err
	}

	return &album, nil
}
