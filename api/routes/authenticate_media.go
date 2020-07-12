package routes

import (
	"database/sql"
	"net/http"

	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"golang.org/x/crypto/bcrypt"
)

func authenticateMedia(media *models.Media, db *sql.DB, r *http.Request) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	user := auth.UserFromContext(r.Context())

	if user != nil {
		row := db.QueryRow("SELECT owner_id FROM album WHERE album.album_id = ?", media.AlbumId)
		var owner_id int

		if err := row.Scan(&owner_id); err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		if owner_id != user.UserID {
			return false, "invalid credentials", http.StatusForbidden, nil
		}
	} else {
		// Check if photo is authorized with a share token
		token := r.URL.Query().Get("token")
		if token == "" {
			return false, "unauthorized", http.StatusForbidden, nil
		}

		row := db.QueryRow("SELECT * FROM share_token WHERE value = ?", token)

		shareToken, err := models.NewShareTokenFromRow(row)
		if err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		// Validate share token password, if set
		if shareToken.Password != nil {
			tokenPassword := r.Header.Get("TokenPassword")

			if err := bcrypt.CompareHashAndPassword([]byte(*shareToken.Password), []byte(tokenPassword)); err != nil {
				if err == bcrypt.ErrMismatchedHashAndPassword {
					return false, "unauthorized", http.StatusForbidden, nil
				} else {
					return false, "internal server error", http.StatusInternalServerError, err
				}
			}
		}

		if shareToken.AlbumID != nil && media.AlbumId != *shareToken.AlbumID {
			// Check child albums
			row := db.QueryRow(`
					WITH recursive child_albums AS (
						SELECT * FROM album WHERE parent_album = ?
						UNION ALL
						SELECT child.* FROM album child JOIN child_albums parent ON parent.album_id = child.parent_album
					)
					SELECT * FROM child_albums WHERE album_id = ?
				`, *shareToken.AlbumID, media.AlbumId)

			_, err := models.NewAlbumFromRow(row)
			if err != nil {
				if err == sql.ErrNoRows {
					return false, "unauthorized", http.StatusForbidden, nil
				}
				return false, "internal server error", http.StatusInternalServerError, err
			}
		}

		if shareToken.MediaID != nil && media.MediaID != *shareToken.MediaID {
			return false, "unauthorized", http.StatusForbidden, nil
		}
	}

	return true, "success", http.StatusAccepted, nil
}
