package resolvers

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userResolver struct {
	*Resolver
}

func (r *Resolver) User() api.UserResolver {
	return &userResolver{r}
}

func (r *queryResolver) User(ctx context.Context, order *models.Ordering, paginate *models.Pagination) ([]*models.User, error) {

	var users []*models.User

	if err := models.FormatSQL(r.Database.Model(models.User{}), order, paginate).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userResolver) Albums(ctx context.Context, user *models.User) ([]*models.Album, error) {
	user.FillAlbums(r.Database)

	pointerAlbums := make([]*models.Album, len(user.Albums))
	for i, album := range user.Albums {
		pointerAlbums[i] = &album
	}

	return pointerAlbums, nil
}

func (r *userResolver) RootAlbums(ctx context.Context, user *models.User) (albums []*models.Album, err error) {

	err = r.Database.Model(&user).
		Where("albums.parent_album_id NOT IN (?)",
			r.Database.Table("user_albums").
				Select("albums.id").
				Joins("JOIN albums ON albums.id = user_albums.album_id AND user_albums.user_id = ?", user.ID),
		).Or("albums.parent_album_id IS NULL").
		Association("Albums").Find(&albums)

	return
}

func (r *queryResolver) MyUser(ctx context.Context) (*models.User, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return user, nil
}

func (r *mutationResolver) AuthorizeUser(ctx context.Context, username string, password string) (*models.AuthorizeResult, error) {
	user, err := models.AuthorizeUser(r.Database, username, password)
	if err != nil {
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	var token *models.AccessToken

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		token, err = user.GenerateAccessToken(tx)
		if err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return &models.AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}

func (r *mutationResolver) InitialSetupWizard(ctx context.Context, username string, password string, rootPath string) (*models.AuthorizeResult, error) {
	siteInfo, err := models.GetSiteInfo(r.Database)
	if err != nil {
		return nil, err
	}

	if !siteInfo.InitialSetup {
		return nil, errors.New("not initial setup")
	}

	rootPath = path.Clean(rootPath)

	var token *models.AccessToken

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE site_info SET initial_setup = false").Error; err != nil {
			return err
		}

		user, err := models.RegisterUser(tx, username, &password, true)
		if err != nil {
			return err
		}

		_, err = scanner.NewRootAlbum(tx, rootPath, user)
		if err != nil {
			return err
		}

		token, err = user.GenerateAccessToken(tx)
		if err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	return &models.AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}

func (r *queryResolver) MyUserPreferences(ctx context.Context) (*models.UserPreferences, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	userPref := models.UserPreferences{
		UserID: user.ID,
	}
	if err := r.Database.Where("user_id = ?", user.ID).FirstOrCreate(&userPref).Error; err != nil {
		return nil, err
	}

	return &userPref, nil
}

func (r *mutationResolver) ChangeUserPreferences(ctx context.Context, language *string) (*models.UserPreferences, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var langTrans *models.LanguageTranslation = nil
	if language != nil {
		lng := models.LanguageTranslation(*language)
		langTrans = &lng
	}

	var userPref models.UserPreferences
	if err := r.Database.Where("user_id = ?", user.ID).FirstOrInit(&userPref).Error; err != nil {
		return nil, err
	}

	userPref.UserID = user.ID
	userPref.Language = langTrans

	if err := r.Database.Save(&userPref).Error; err != nil {
		return nil, err
	}

	return &userPref, nil
}

// Admin queries
func (r *mutationResolver) UpdateUser(ctx context.Context, id int, username *string, password *string, admin *bool) (*models.User, error) {

	if username == nil && password == nil && admin == nil {
		return nil, errors.New("no updates requested")
	}

	var user models.User
	if err := r.Database.First(&user, id).Error; err != nil {
		return nil, err
	}

	if username != nil {
		user.Username = *username
	}

	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, err
		}
		hashedPass := string(hashedPassBytes)

		user.Password = &hashedPass
	}

	if admin != nil {
		user.Admin = *admin
	}

	if err := r.Database.Save(&user).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update user")
	}

	return &user, nil
}

func (r *mutationResolver) CreateUser(ctx context.Context, username string, password *string, admin bool) (*models.User, error) {

	var user *models.User

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		var err error
		user, err = models.RegisterUser(tx, username, password, admin)
		if err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return user, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	deletedAlbumIDs := make([]int, 0)

	err := r.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, id).Error; err != nil {
			return err
		}

		userAlbums := user.Albums
		if err := tx.Model(&user).Association("Albums").Find(&userAlbums); err != nil {
			return err
		}

		if err := tx.Model(&user).Association("Albums").Clear(); err != nil {
			return err
		}

		for _, album := range userAlbums {
			var associatedUsers = tx.Model(album).Association("Owners").Count()

			if associatedUsers == 0 {
				deletedAlbumIDs = append(deletedAlbumIDs, album.ID)
				if err := tx.Delete(album).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If there is only one associated user, clean up the cache folder and delete the album row
	for _, deletedAlbumID := range deletedAlbumIDs {
		cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(deletedAlbumID)))
		if err := os.RemoveAll(cachePath); err != nil {
			return &user, err
		}
	}
	return &user, nil
}

func (r *mutationResolver) UserAddRootPath(ctx context.Context, id int, rootPath string) (*models.Album, error) {

	rootPath = path.Clean(rootPath)

	var user models.User
	if err := r.Database.First(&user, id).Error; err != nil {
		return nil, err
	}

	if !models.ValidRootPath(rootPath) {
		return nil, errors.New("invalid root path")
	}

	upperPaths := make([]string, 1)
	upperPath := rootPath
	upperPaths[0] = upperPath
	for {

		substrIndex := strings.LastIndex(upperPath, "/")
		if substrIndex == -1 {
			break
		}

		if substrIndex == 0 {
			upperPaths = append(upperPaths, "/")
			break
		}

		upperPath = upperPath[0:substrIndex]
		upperPaths = append(upperPaths, upperPath)
	}

	var upperAlbums []models.Album
	if err := r.Database.Model(&user).Association("Albums").Find(&upperAlbums, "albums.path IN (?)", upperPaths); err != nil {
		// if err := r.Database.Model(models.Album{}).Where("path IN (?)", upperPaths).Find(&upperAlbums).Error; err != nil {
		return nil, err
	}

	if len(upperAlbums) > 0 {
		return nil, errors.New(fmt.Sprintf("user already owns a path containing this path: %s", upperAlbums[0].Path))
	}

	newAlbum, err := scanner.NewRootAlbum(r.Database, rootPath, &user)
	if err != nil {
		return nil, err
	}

	return newAlbum, nil
}

func (r *mutationResolver) UserRemoveRootAlbum(ctx context.Context, userID int, albumID int) (*models.Album, error) {

	var album models.Album
	if err := r.Database.First(&album, albumID).Error; err != nil {
		return nil, err
	}

	var deletedAlbumIDs []int = nil

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw("DELETE FROM user_albums WHERE user_id = ? AND album_id = ?", userID, albumID).Error; err != nil {
			return err
		}

		children, err := album.GetChildren(tx, nil)
		if err != nil {
			return err
		}

		childAlbumIDs := make([]int, len(children))
		for i, child := range children {
			childAlbumIDs[i] = child.ID
		}

		result := tx.Exec("DELETE FROM user_albums WHERE user_id = ? and album_id IN (?)", userID, childAlbumIDs)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("No relation deleted")
		}

		// Cleanup if no user owns the album anymore
		var userAlbumCount int
		if err := tx.Raw("SELECT COUNT(user_id) FROM user_albums WHERE album_id = ?", albumID).Scan(&userAlbumCount).Error; err != nil {
			return err
		}

		if userAlbumCount == 0 {
			deletedAlbumIDs = append(childAlbumIDs, albumID)
			childAlbumIDs = nil

			// Delete albums from database
			if err := tx.Delete(&models.Album{}, "id IN (?)", deletedAlbumIDs).Error; err != nil {
				deletedAlbumIDs = nil
				return err
			}
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	if deletedAlbumIDs != nil {
		// Delete albums from cache
		for _, id := range deletedAlbumIDs {
			cacheAlbumPath := path.Join(utils.MediaCachePath(), strconv.Itoa(id))

			if err := os.RemoveAll(cacheAlbumPath); err != nil {
				return nil, err
			}
		}

		// Reload faces as media might have been deleted
		if err := face_detection.GlobalFaceDetector.ReloadFacesFromDatabase(r.Database); err != nil {
			return nil, err
		}
	}

	return &album, nil
}
