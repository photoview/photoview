package routes

import (
	"fmt"
	"net/http"

	"github.com/kkovaletp/photoview/api/graphql/auth"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func authenticateMedia(media *models.Media, db *gorm.DB, r *http.Request) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	user := auth.UserFromContext(r.Context())

	if user != nil {
		var album models.Album
		if err := db.First(&album, media.AlbumID).Error; err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		ownsAlbum, err := user.OwnsAlbum(db, &album)
		if err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		if !ownsAlbum {
			return false, "invalid credentials", http.StatusForbidden, nil
		}
	} else {
		if success, respMsg, respStatus, err := shareTokenFromRequest(db, r, &media.ID, &media.AlbumID); !success {
			return success, respMsg, respStatus, err
		}

	}

	return true, "success", http.StatusAccepted, nil
}

func authenticateAlbum(album *models.Album, db *gorm.DB, r *http.Request) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	user := auth.UserFromContext(r.Context())

	if user != nil {
		ownsAlbum, err := user.OwnsAlbum(db, album)
		if err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		if !ownsAlbum {
			return false, "invalid credentials", http.StatusForbidden, nil
		}
	} else {
		if success, respMsg, respStatus, err := shareTokenFromRequest(db, r, nil, &album.ID); !success {
			return success, respMsg, respStatus, err
		}

	}

	return true, "success", http.StatusAccepted, nil
}

func shareTokenFromRequest(db *gorm.DB, r *http.Request, mediaID *int, albumID *int) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	// Check if photo is authorized with a share token
	token := r.URL.Query().Get("token")
	if token == "" {
		return false, "unauthorized", http.StatusForbidden, errors.New("share token not provided")
	}

	var shareToken models.ShareToken

	if err := db.Where("value = ?", token).First(&shareToken).Error; err != nil {
		return false, "internal server error", http.StatusInternalServerError, err
	}

	// Validate share token password, if set
	if shareToken.Password != nil {
		tokenPasswordCookie, err := r.Cookie(fmt.Sprintf("share-token-pw-%s", shareToken.Value))
		if err != nil {
			return false, "unauthorized", http.StatusForbidden, errors.Wrap(err, "get share token password cookie")
		}
		// tokenPassword := r.Header.Get("TokenPassword")
		tokenPassword := tokenPasswordCookie.Value

		if err := bcrypt.CompareHashAndPassword([]byte(*shareToken.Password), []byte(tokenPassword)); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return false, "unauthorized", http.StatusForbidden, errors.New("incorrect password for share token")
			} else {
				return false, "internal server error", http.StatusInternalServerError, err
			}
		}
	}

	if shareToken.AlbumID != nil && albumID == nil {
		return false, "unauthorized", http.StatusForbidden, errors.New("share token is of type album, but no albumID was provided to function")
	}

	if shareToken.MediaID != nil && mediaID == nil {
		return false, "unauthorized", http.StatusForbidden, errors.New("share token is of type media, but no mediaID was provided to function")
	}

	if shareToken.AlbumID != nil && *albumID != *shareToken.AlbumID {
		// Check child albums

		var count int
		err := db.Raw(`
				WITH recursive child_albums AS (
					SELECT * FROM albums WHERE parent_album_id = ?
					UNION ALL
					SELECT child.* FROM albums child JOIN child_albums parent ON parent.id = child.parent_album_id
				)
				SELECT COUNT(id) FROM child_albums WHERE id = ?
			`, *shareToken.AlbumID, albumID).Find(&count).Error

		if err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		if count == 0 {
			return false, "unauthorized", http.StatusForbidden, errors.New("no child albums found for share token")
		}
	}

	if shareToken.MediaID != nil && *mediaID != *shareToken.MediaID {
		return false, "unauthorized", http.StatusForbidden, errors.New("media share token does not match mediaID")
	}

	return true, "", 0, nil
}
