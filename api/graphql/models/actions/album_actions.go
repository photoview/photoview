package actions

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func MyAlbums(db *gorm.DB, user *models.User, order *models.Ordering, paginate *models.Pagination,
	onlyRoot *bool, showEmpty *bool, onlyWithFavorites *bool) ([]*models.Album, error) {

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

	query = models.FormatSQL(query, order, paginate)

	var albums []*models.Album
	if err := query.Find(&albums).Error; err != nil {
		return nil, err
	}

	return albums, nil
}

func getSingleRootAlbumID(user *models.User) int {
	var singleRootAlbumID int = -1
	for _, album := range user.Albums {
		if album.ParentAlbumID == nil {
			if singleRootAlbumID == -1 {
				singleRootAlbumID = album.ID
			} else {
				singleRootAlbumID = -1
				break
			}
		}
	}
	return singleRootAlbumID
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

	ownsAlbum, err := user.OwnsAlbum(db, &album)
	if err != nil {
		return nil, err
	}

	if !ownsAlbum {
		return nil, errors.New("forbidden")
	}

	return &album, nil
}

func AlbumPath(db *gorm.DB, user *models.User, album *models.Album) ([]*models.Album, error) {
	var albumPath []*models.Album

	// depth is carried through the recursion and used only to ORDER BY - a
	// recursive CTE gives no ordering guarantee otherwise, and the
	// truncation below depends on seeing the closest ancestor first.
	//
	// It also bounds the recursion. Without depth, UNION alone ended a cyclic
	// parent chain by deduplicating the repeated rows; with it every lap
	// produces a distinct row, so a cycle would run until the database gives
	// up. Nothing creates such a cycle today - parents come from the
	// directory tree - but the query shouldn't be the thing that turns a bad
	// row into a hang. No real album tree is anywhere near this deep.
	if err := db.Raw(`
		WITH recursive path_albums AS (
			SELECT *, 0 AS depth FROM albums anchor WHERE anchor.id = ?
			UNION
			SELECT parent.*, child.depth + 1 FROM path_albums child
				JOIN albums parent ON parent.id = child.parent_album_id
				WHERE child.depth < 64
		)
		SELECT * FROM path_albums WHERE id != ? ORDER BY depth ASC
	`, album.ID, album.ID).Scan(&albumPath).Error; err != nil {
		return nil, err
	}

	// albumPath runs closest-ancestor-first, root-last. Walk outward from
	// the album and stop at the first ancestor the user doesn't own:
	// everything from there to the root is dropped, while the closer
	// ancestors they do own are kept. Truncating from the root end instead
	// would throw away the whole breadcrumb as soon as the top of the tree
	// happens to be someone else's.
	visibleUpTo := len(albumPath)

	for i, ancestor := range albumPath {
		owns, err := user.OwnsAlbum(db, ancestor)
		if err != nil {
			return nil, err
		}

		if !owns {
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

	ownsAlbum, err := user.OwnsAlbum(db, &album)
	if err != nil {
		return nil, err
	}

	if !ownsAlbum {
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

	ownsAlbum, err := user.OwnsAlbum(db, &album)
	if err != nil {
		return nil, err
	}

	if !ownsAlbum {
		return nil, errors.New("forbidden")
	}

	if err := db.Model(&album).Update("cover_id", nil).Error; err != nil {
		return nil, err
	}

	return &album, nil
}
