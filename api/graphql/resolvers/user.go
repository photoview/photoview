package resolvers

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
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

	if err := models.FormatSQL(r.DB(ctx).Model(models.User{}), order, paginate).Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}
	for _, user := range users {
		user.Admin = user.Role.Name == "ADMIN"
	}
	return users, nil
}

func (r *userResolver) Albums(ctx context.Context, user *models.User) ([]*models.Album, error) {
	user.FillAlbums(r.DB(ctx))

	pointerAlbums := make([]*models.Album, len(user.Albums))
	for i, album := range user.Albums {
		pointerAlbums[i] = &album
	}

	return pointerAlbums, nil
}

func (r *userResolver) RootAlbums(ctx context.Context, user *models.User) (albums []*models.Album, err error) {
	db := r.DB(ctx)

	err = db.Model(&user).
		Where("albums.parent_album_id NOT IN (?)",
			db.Table("user_albums").
				Select("albums.id").
				Joins("JOIN albums ON albums.id = user_albums.album_id AND user_albums.user_id = ?", user.ID),
		).Or("albums.parent_album_id IS NULL").
		Association("Albums").Find(&albums)

	return
}

func (r *queryResolver) Roles(ctx context.Context) ([]*models.Role, error) {
	db := r.DB(ctx)
	results := make([]*models.Role, 0)
	db.Preload("Permissions").Find(&results)
	return results, nil
}
func (r *queryResolver) Permissions(ctx context.Context) ([]*models.PermissionModel, error) {
	db := r.DB(ctx)
	results := make([]*models.PermissionModel, 0)
	db.Find(&results)
	return results, nil
}

func (r *queryResolver) MyUser(ctx context.Context) (*models.User, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return user, nil
}

func (r *mutationResolver) CreateRole(ctx context.Context, role *models.NewRoleInput) (*models.Role, error) {
	db := r.DB(ctx)
	permissions := make([]*models.PermissionModel, 0, len(role.Permissions))
	db.Find(&permissions, role.Permissions)

	newRole := &models.Role{
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
	}
	db.Save(newRole)
	return newRole, nil
}

func (r *mutationResolver) DeleteRole(ctx context.Context, id int) (*models.Role, error) {
	db := r.DB(ctx)
	role := &models.Role{}
	db.Find(role, id)
	if role.SystemRole {
		return nil, errors.New("Unable to delete system roles")
	}

	idsUsingRole := make([]int, 0)
	db.Model(&models.User{}).Where("roleId = ?", id).Pluck("id", &idsUsingRole)

	if len(idsUsingRole) != 0 {
		return nil, errors.New("There are still users assigned this role")
	}
	db.Delete(role)
	return role, nil
}

func (r *mutationResolver) UpdateRole(ctx context.Context, role *models.UpdateRoleInput) (*models.Role, error) {
	db := r.DB(ctx)
	currentRole := &models.Role{}
	db.Find(currentRole, role.ID)
	if !currentRole.Editable {
		return nil, fmt.Errorf("The role cannot be edited")
	}
	if role.Name != nil && !currentRole.SystemRole {
		currentRole.Name = *role.Name
	} else if currentRole.SystemRole {
		return nil, fmt.Errorf("Cannoy update the name of a system role")
	}
	if role.Description != nil {
		currentRole.Description = *role.Description
	}
	if role.Permissions != nil {
		if len(role.Permissions) == 0 {
			currentRole.Permissions = make([]*models.PermissionModel, 0)
		} else {
			permissions := make([]*models.PermissionModel, 0, len(role.Permissions))
			db.Find(&permissions, role.Permissions)
			currentRole.Permissions = permissions
		}
	}

	db.Session(&gorm.Session{FullSaveAssociations: true}).Save(currentRole)
	return currentRole, nil
}

func (r *mutationResolver) AuthorizeUser(ctx context.Context, username string, password string) (*models.AuthorizeResult, error) {
	db := r.DB(ctx)
	user, err := models.AuthorizeUser(db, username, password)
	if err != nil {
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	var token *models.AccessToken

	transactionError := db.Transaction(func(tx *gorm.DB) error {
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
	db := r.DB(ctx)
	siteInfo, err := models.GetSiteInfo(db)
	if err != nil {
		return nil, err
	}

	if !siteInfo.InitialSetup {
		return nil, errors.New("not initial setup")
	}

	rootPath = path.Clean(rootPath)

	var token *models.AccessToken

	transactionError := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE site_info SET initial_setup = false").Error; err != nil {
			return err
		}
		ids := make([]int, 0)

		r.DB(ctx).Model(&models.Role{}).Where("name = ?", "ADMIN").Pluck("id", &ids)
		// TODO init roles here. then pass saved role to user.
		user, err := models.RegisterUser(tx, username, &password, ids[0])
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
	if err := r.DB(ctx).Where("user_id = ?", user.ID).FirstOrCreate(&userPref).Error; err != nil {
		return nil, err
	}

	return &userPref, nil
}

func (r *mutationResolver) ChangeUserPreferences(ctx context.Context, language *string) (*models.UserPreferences, error) {
	db := r.DB(ctx)
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
	if err := db.Where("user_id = ?", user.ID).FirstOrInit(&userPref).Error; err != nil {
		return nil, err
	}

	userPref.UserID = user.ID
	userPref.Language = langTrans

	if err := db.Save(&userPref).Error; err != nil {
		return nil, err
	}

	return &userPref, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id int, username *string, password *string, roleID *int, admin *bool) (*models.User, error) {
	db := r.DB(ctx)

	if username == nil && password == nil && admin == nil && roleID == nil {
		return nil, errors.New("no updates requested")
	}

	if admin != nil && roleID != nil {
		return nil, errors.New("Cannot use roleId and admin properties together")

	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}

	// TODO it's currently possible to remove the last admin this way we need a check in place

	if admin != nil || roleID != nil {
		realRoleId, err := resolveRealRoleId(ctx, roleID, admin, r)
		if err != nil {
			return nil, errors.New("unable to identify role change id")
		}
		adminIds := make([]int, 0)
		db.Model(&models.Role{}).Where("name = ?", "ADMIN").Pluck("id", &adminIds)
		if len(adminIds) == 1 && adminIds[0] == user.ID && *user.RoleID != realRoleId {
			return nil, errors.New("Cannot remove the last admin")
		}
		user.RoleID = &realRoleId
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

	if err := db.Save(&user).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update user")
	}

	return &user, nil
}

func (r *mutationResolver) CreateUser(ctx context.Context, username string, password *string, roleId *int, admin *bool) (*models.User, error) {
	var user *models.User
	realRoleId, err := resolveRealRoleId(ctx, roleId, admin, r)
	if err != nil {
		return nil, err
	}
	transactionError := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		user, err = models.RegisterUser(tx, username, password, realRoleId)
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

func resolveRealRoleId(ctx context.Context, roleId *int, admin *bool, r *mutationResolver) (int, error) {
	var realRoleId = -1
	if roleId == nil && admin == nil {
		return 0, fmt.Errorf("expecting one of roleId || admin to be defined")
	}
	if roleId != nil {
		realRoleId = *roleId
	} else {
		var ids []int
		if *admin {
			r.DB(ctx).Model(&models.Role{}).Where("name = ?", "ADMIN").Pluck("id", &ids)
		} else {
			r.DB(ctx).Model(&models.Role{}).Where("name = ?", "USER").Pluck("id", &ids)
		}

		if len(ids) != 1 {
			return 0, fmt.Errorf("expecting 1 role id to be returned recieved %d", len(ids))
		}
		realRoleId = ids[0]
	}
	return realRoleId, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id int) (*models.User, error) {
	return actions.DeleteUser(r.DB(ctx), id)
}

func (r *mutationResolver) UserAddRootPath(ctx context.Context, id int, rootPath string) (*models.Album, error) {
	db := r.DB(ctx)

	rootPath = path.Clean(rootPath)

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}

	newAlbum, err := scanner.NewRootAlbum(db, rootPath, &user)
	if err != nil {
		return nil, err
	}

	return newAlbum, nil
}

func (r *mutationResolver) UserRemoveRootAlbum(ctx context.Context, userID int, albumID int) (*models.Album, error) {
	db := r.DB(ctx)

	var album models.Album
	if err := db.First(&album, albumID).Error; err != nil {
		return nil, err
	}

	var deletedAlbumIDs []int = nil

	transactionError := db.Transaction(func(tx *gorm.DB) error {
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
		if face_detection.GlobalFaceDetector != nil {
			if err := face_detection.GlobalFaceDetector.ReloadFacesFromDatabase(db); err != nil {
				return nil, err
			}
		}
	}

	return &album, nil
}
