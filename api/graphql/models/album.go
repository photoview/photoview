package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Album struct {
	Model
	Title         string `gorm:"not null"`
	ParentAlbumID *int   `gorm:"index"`
	ParentAlbum   *Album `gorm:"constraint:OnDelete:SET NULL;"`
	// OwnerID       int `gorm:"not null"`
	// Owner         User
	Owners   []User `gorm:"many2many:user_albums;constraint:OnDelete:CASCADE;"`
	Path     string `gorm:"not null"`
	PathHash string `gorm:"unique"`
	CoverID  *int
}

func (a *Album) FilePath() string {
	return a.Path
}

func (a *Album) BeforeSave(tx *gorm.DB) (err error) {
	hash := md5.Sum([]byte(a.Path))
	a.PathHash = hex.EncodeToString(hash[:])
	return nil
}

// GetChildren performs a recursive query to get all the children of the album.
// An optional filter can be provided that can be used to modify the query on the children.
func (a *Album) GetChildren(db *gorm.DB, filter func(*gorm.DB) *gorm.DB) (children []*Album, err error) {
	return GetChildrenFromAlbums(db, filter, []int{a.ID})
}

func GetChildrenFromAlbums(db *gorm.DB, filter func(*gorm.DB) *gorm.DB, albumIDs []int) (children []*Album, err error) {
	query := db.Model(&Album{}).Table("sub_albums")

	if filter != nil {
		query = filter(query)
	}

	err = db.Raw(`
	WITH recursive sub_albums AS (
		SELECT * FROM albums AS root WHERE id IN (?)
		UNION ALL
		SELECT child.* FROM albums AS child JOIN sub_albums ON child.parent_album_id = sub_albums.id
	)

	?
	`, albumIDs, query).Find(&children).Error

	return children, err
}

func (a *Album) GetParents(db *gorm.DB, filter func(*gorm.DB) *gorm.DB) (parents []*Album, err error) {
	return GetParentsFromAlbums(db, filter, a.ID)
}

func GetParentsFromAlbums(db *gorm.DB, filter func(*gorm.DB) *gorm.DB, albumID int) (parents []*Album, err error) {
	query := db.Model(&Album{}).Table("super_albums")

	if filter != nil {
		query = filter(query)
	}

	err = db.Raw(`
	WITH recursive super_albums AS (
		SELECT * FROM albums AS leaf WHERE id = ?
		UNION ALL
		SELECT parent.* from albums AS parent JOIN super_albums ON parent.id = super_albums.parent_album_id
	)

	?
	`, albumID, query).Find(&parents).Error

	return parents, err
}

// PropagateAlbumLevel stamps level (and grantedByUserID) onto albumID and
// every one of its *current* descendants for userID, as a grant sourced at
// albumID. This is needed any time a grant is created or its level changes
// on an album that may already have content scanned in below it: the
// scanner's own copy-owners-onto-new-child step only ever reaches
// directories discovered *after* the grant exists, so a backlog of
// already-scanned descendants would otherwise keep a stale (or missing)
// level.
//
// The write goes through UserAlbumGrant, keyed on (user, album, source) so
// a grant reaching the same descendant from a *different* source (e.g. an
// admin's root grant reaching a folder an owner already shared separately)
// coexists instead of overwriting it - UserAlbums.Level ends up the max
// across every source, recomputed by recomputeUserAlbums.
func PropagateAlbumLevel(db *gorm.DB, albumID int, userID int, level AlbumPermissionLevel, grantedByUserID *int) error {
	var album Album
	if err := db.First(&album, albumID).Error; err != nil {
		return err
	}

	subtree, err := album.GetChildren(db, nil)
	if err != nil {
		return err
	}

	grants := make([]UserAlbumGrant, len(subtree))
	albumIDs := make([]int, len(subtree))
	for i, a := range subtree {
		grants[i] = UserAlbumGrant{UserID: userID, AlbumID: a.ID, SourceAlbumID: albumID, Level: level, GrantedByUserID: grantedByUserID}
		albumIDs[i] = a.ID
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "album_id"}, {Name: "source_album_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"level", "granted_by_user_id"}),
	}).Create(&grants).Error; err != nil {
		return err
	}

	return recomputeUserAlbums(db, userID, albumIDs)
}

// RevokeAlbumLevel removes userID's access to albumID and every one of its
// current descendants that was sourced at albumID - i.e. exactly what the
// matching PropagateAlbumLevel(albumID, userID, ...) call created. A grant
// reaching the same albums from a different source is untouched, since it's
// a different UserAlbumGrant row by construction.
func RevokeAlbumLevel(db *gorm.DB, albumID int, userID int) error {
	var album Album
	if err := db.First(&album, albumID).Error; err != nil {
		return err
	}

	subtree, err := album.GetChildren(db, nil)
	if err != nil {
		return err
	}

	ids := make([]int, len(subtree))
	for i, a := range subtree {
		ids[i] = a.ID
	}

	if err := db.Where("user_id = ? AND album_id IN (?) AND source_album_id = ?", userID, ids, albumID).
		Delete(&UserAlbumGrant{}).Error; err != nil {
		return err
	}

	return recomputeUserAlbums(db, userID, ids)
}

// recomputeUserAlbums recalculates the materialized UserAlbums row for each
// (userID, albumID) pair in albumIDs from the current UserAlbumGrant rows:
// the max Level across every source, and a nil GrantedByUserID (owner
// access) if any source is owner-rooted, else an arbitrary non-nil
// grantor. An album with no remaining grant rows has its UserAlbums row
// deleted, matching the "no access" semantics of a full revoke.
func recomputeUserAlbums(db *gorm.DB, userID int, albumIDs []int) error {
	if len(albumIDs) == 0 {
		return nil
	}

	var grants []UserAlbumGrant
	if err := db.Where("user_id = ? AND album_id IN (?)", userID, albumIDs).Find(&grants).Error; err != nil {
		return err
	}

	byAlbum := make(map[int][]UserAlbumGrant, len(albumIDs))
	for _, g := range grants {
		byAlbum[g.AlbumID] = append(byAlbum[g.AlbumID], g)
	}

	toUpsert := make([]UserAlbums, 0, len(albumIDs))
	toDelete := make([]int, 0, len(albumIDs))
	for _, albumID := range albumIDs {
		sources := byAlbum[albumID]
		if len(sources) == 0 {
			toDelete = append(toDelete, albumID)
			continue
		}

		best := sources[0]
		grantedBy := best.GrantedByUserID
		for _, g := range sources[1:] {
			if g.Level.HigherThan(best.Level) {
				best = g
			}
			if g.GrantedByUserID == nil {
				grantedBy = nil
			}
		}

		toUpsert = append(toUpsert, UserAlbums{UserID: userID, AlbumID: albumID, Level: best.Level, GrantedByUserID: grantedBy})
	}

	if len(toUpsert) > 0 {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "album_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"level", "granted_by_user_id"}),
		}).Create(&toUpsert).Error; err != nil {
			return err
		}
	}

	if len(toDelete) > 0 {
		if err := db.Where("user_id = ? AND album_id IN (?)", userID, toDelete).Delete(&UserAlbums{}).Error; err != nil {
			return err
		}
	}

	return nil
}

// HiddenAlbumsClosure returns the ids of every album that should be treated
// as hidden for userID in search results: the albums they've explicitly
// hidden, plus all descendants of those albums (a descendant isn't itself
// marked hidden, but is unreachable by browsing once its ancestor is
// hidden, so it shouldn't surface via search either).
func HiddenAlbumsClosure(db *gorm.DB, userID int) ([]int, error) {
	var hiddenAlbumIDs []int
	if err := db.Model(&UserAlbumData{}).
		Where("user_id = ? AND hidden = true", userID).
		Pluck("album_id", &hiddenAlbumIDs).Error; err != nil {
		return nil, err
	}

	if len(hiddenAlbumIDs) == 0 {
		return nil, nil
	}

	closureAlbums, err := GetChildrenFromAlbums(db, nil, hiddenAlbumIDs)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(closureAlbums))
	for i, a := range closureAlbums {
		ids[i] = a.ID
	}
	return ids, nil
}

func (a *Album) Thumbnail(db *gorm.DB) (*Media, error) {
	var media Media

	if a.CoverID != nil {
		if err := db.First(&media, *a.CoverID).Error; err != nil {
			return nil, err
		}
		return &media, nil
	}

	query := `
		WITH RECURSIVE sub_albums AS (
			SELECT id FROM albums WHERE id = ?
			UNION ALL
			SELECT children.id FROM albums AS children
			INNER JOIN sub_albums ON children.parent_album_id = sub_albums.id
		)
		SELECT * FROM media
		WHERE media.album_id IN (SELECT id FROM sub_albums)
		ORDER BY media.id DESC
		LIMIT 1
	`

	if err := db.Raw(query, a.ID).Scan(&media).Error; err != nil {
		return nil, err
	}

	if media.ID == 0 {
		return nil, nil // Return nil for empty albums
	}

	return &media, nil
}
