package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"

	// "github.com/photoview/photoview/api/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const internalServerError = "internal server error"

func authenticateMedia(media *models.Media, db *gorm.DB, r *http.Request) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	user := auth.UserFromContext(r.Context())

	if user != nil {
		var album models.Album
		if err := db.First(&album, media.AlbumID).Error; err != nil {
			// log.Debug(nil, "Failed to find album for media %d: %v", media.ID, err)
			return false, internalServerError, http.StatusInternalServerError, err
		}

		ownsAlbum, err := user.OwnsAlbum(db, &album)
		if err != nil {
			// log.Debug(nil, "Failed to check if user owns album %d for media %d: %v", media.AlbumID, media.ID, err)
			return false, internalServerError, http.StatusInternalServerError, err
		}

		if !ownsAlbum {
			// log.Debug(nil, "User does not own album %d for media %d", media.AlbumID, media.ID)
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
			// log.Debug(nil, "Failed to check if user owns album %d: %v", album.ID, err)
			return false, internalServerError, http.StatusInternalServerError, err
		}

		if !ownsAlbum {
			// log.Debug(nil, "User does not own album %d", album.ID)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// log.Debug(nil, "Share token not found: %s", token)
			return false, "unauthorized", http.StatusForbidden, errors.New("invalid share token")
		}
		// log.Debug(nil, "Error fetching share token: %s, error: %v", token, err)
		return false, internalServerError, http.StatusInternalServerError, err
	}

	if shareToken.Expire != nil && time.Now().UTC().After(shareToken.Expire.UTC()) {
		// log.Debug(nil, "Share token expired: %s", token)
		return false, "unauthorized", http.StatusForbidden, errors.New("invalid share token")
	}

	// Validate share token password, if set
	if shareToken.Password != nil {
		tokenPasswordCookie, err := r.Cookie(fmt.Sprintf("share-token-pw-%s", shareToken.Value))
		if err != nil {
			// log.Debug(nil, "Error getting share token password cookie: %v", err)
			return false, "unauthorized", http.StatusForbidden, errors.Wrap(err, "share token password invalid")
		}
		// tokenPassword := r.Header.Get("TokenPassword")
		tokenPassword := tokenPasswordCookie.Value

		if err := bcrypt.CompareHashAndPassword([]byte(*shareToken.Password), []byte(tokenPassword)); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				// log.Debug(nil, "Incorrect password for share token: %s", token)
				return false, "unauthorized", http.StatusForbidden, errors.New("share token password invalid")
			} else {
				// log.Debug(nil, "Error comparing share token password: %s, error: %v", token, err)
				return false, internalServerError, http.StatusInternalServerError, err
			}
		}
	}

	if shareToken.AlbumID != nil && albumID == nil {
		// log.Debug(nil, "Share token is of type album, but no albumID was provided to function")
		return false, "unauthorized", http.StatusForbidden, errors.New("invalid share token")
	}

	if shareToken.MediaID != nil && mediaID == nil {
		// log.Debug(nil, "Share token is of type media, but no mediaID was provided to function")
		return false, "unauthorized", http.StatusForbidden, errors.New("invalid share token")
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
			// log.Debug(nil, "Error checking child albums for share token: %s, error: %v", token, err)
			return false, internalServerError, http.StatusInternalServerError, err
		}

		if count == 0 {
			// log.Debug(nil, "No child albums found for share token: %s", token)
			return false, "unauthorized", http.StatusForbidden, errors.New("invalid share token")
		}
	}

	if shareToken.MediaID != nil && *mediaID != *shareToken.MediaID {
		// log.Debug(nil, "Media share token does not match mediaID: %d != %d", *mediaID, *shareToken.MediaID)
		return false, "unauthorized", http.StatusForbidden, errors.New("invalid share token")
	}

	return true, "", 0, nil
}
