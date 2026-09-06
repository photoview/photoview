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
	query = HiddenAlbumsFilter(showHidden, db, user, query)

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

// hiddenAlbumsFilter excludes albums the user has personally hidden, unless
// showHidden is true (used to reveal hidden albums, dimmed, in the UI).
func HiddenAlbumsFilter(showHidden *bool, db *gorm.DB, user *models.User, query *gorm.DB) *gorm.DB {
	if showHidden != nil && *showHidden {
		return query
	}
	hiddenSubquery := db.Model(&models.UserAlbumData{UserID: user.ID}).
		Where("user_album_data.album_id = albums.id").
		Where("user_album_data.hidden = true")
	return query.Where("NOT EXISTS (?)", hiddenSubquery)
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

	err := db.Raw(`
		WITH recursive path_albums AS (
			SELECT * FROM albums anchor WHERE anchor.id = ?
			UNION
			SELECT parent.* FROM path_albums child JOIN albums parent ON parent.id = child.parent_album_id
		)
		SELECT * FROM path_albums WHERE id != ?
	`, album.ID, album.ID).Scan(&albumPath).Error

	// Truncate the path at the point the user can no longer see, e.g. when
	// they were only granted a subfolder rather than one of its ancestors.
	for i := len(albumPath) - 1; i >= 0; i-- {
		album := albumPath[i]

		hasAccess, err := user.HasAlbumLevel(db, album, models.AlbumPermissionLevelRead)
		if err != nil {
			return nil, err
		}

		if !hasAccess {
			albumPath = albumPath[i+1:]
			break
		}

	}

	if err != nil {
		return nil, err
	}

	return albumPath, nil
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

	hasAccess, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelRead)
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

	hasAccess, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelRead)
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
