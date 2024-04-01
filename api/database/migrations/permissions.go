package migrations

import (
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// MigrateForPermissions Migrates new permissions into the system and upgrades system roles according to the new permissions.
func MigrateForPermissions(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		newPermissions, err := migrateNewPermissions(tx)
		if err != nil {
			return err
		}
		if err := migrateOrCreateSystemRoles(tx, newPermissions); err != nil {
			return err
		}

		initialMigration(tx)
		return tx.Error
	})
}

func initialMigration(db *gorm.DB) {
	if db.Migrator().HasColumn(&models.User{}, "admin") {
		migrateExistingAdminUsers(db)
		migrateExistingUsers(db)
		migrateDemoUser(db)
		db.Migrator().DropColumn(&models.User{}, "admin")
	}
}

func migrateOrCreateSystemRoles(db *gorm.DB, newPermissions []*models.PermissionModel) error {
	if err := migrateOrCreateSystemRole(db, "ADMIN", false, make([]models.Permission, 0), make([]*models.PermissionModel, 0)); err != nil {
		return err
	}
	if err := migrateOrCreateSystemRole(db, "DEMO", true, auth.DEMO, newPermissions); err != nil {
		return err
	}
	if err := migrateOrCreateSystemRole(db, "USER", true, auth.USER, newPermissions); err != nil {
		return err
	}
	return nil
}

func migrateOrCreateSystemRole(db *gorm.DB, roleName string, editable bool, expectedPermissions []models.Permission, newPermissions []*models.PermissionModel) error {
	// initilise or find existing role.
	newRole := models.Role{Name: roleName, SystemRole: true, Editable: editable, Permissions: make([]*models.PermissionModel, 0)}
	result := db.Where("name = ?", roleName).FirstOrCreate(&newRole)

	if result.Error != nil {
		return result.Error
	}
	// we want to return here if we're not expecting any permissions or we are but have just created the role.
	if result.RowsAffected != 0 || len(expectedPermissions) == 0 {
		// Initilise base roles if we expecting permissions.
		if len(expectedPermissions) == 0 {
			db.Find(&newRole.Permissions).Where("name in (?)", expectedPermissions)
			if err := db.Save(&newRole).Error; err != nil {
				return nil
			}
		}
		return nil
	}

	releventPermissions := findReleventPermissions(newPermissions, expectedPermissions)
	if len(releventPermissions) == 0 {
		return nil
	}
	err := updateRole(db, &newRole, releventPermissions)
	if err != nil {
		return err
	}

	return nil
}

func findReleventPermissions(newPermissions []*models.PermissionModel, expectedRoles []models.Permission) []*models.PermissionModel {
	expectedMap := make(map[models.Permission]bool)
	for _, name := range expectedRoles {
		expectedMap[name] = true
	}
	releventPermissions := make([]*models.PermissionModel, 0)
	for _, permission := range newPermissions {
		if expectedMap[permission.Name] {
			releventPermissions = append(releventPermissions, permission)
		}
	}
	return releventPermissions
}

func updateRole(db *gorm.DB, role *models.Role, releventPermissions []*models.PermissionModel) error {
	role.Permissions = append(make([]*models.PermissionModel, 0), releventPermissions...)
	if err := db.Save(role).Error; err != nil {
		return err
	}
	return nil
}

func migrateExistingAdminUsers(db *gorm.DB) {
	currentAdmins := make([]*models.User, 0)

	db.Where("admin = ?", true).Find(&currentAdmins)
	for _, admin := range currentAdmins {
		db.Where("name = ?", "ADMIN").First(&admin.Role)
	}
	db.Save(&currentAdmins)
}

func migrateExistingUsers(db *gorm.DB) error {
	currentUsers := make([]*models.User, 0)

	db.Where("role_id is null").Find(&currentUsers)
	for _, user := range currentUsers {
		db.Where("name = ?", "USER").First(&user.Role)
	}
	if err := db.Save(&currentUsers).Error; err != nil {
		return errors.Wrap(err, "failed to migrate existing users to USER role")
	}
	return nil
}

func migrateDemoUser(db *gorm.DB) error {
	demoUser := &models.User{}
	results := db.Where("username = ?", "demo").Find(demoUser).RowsAffected
	if results == 0 {
		return nil
	}
	db.Where("name = ?", "DEMO").First(&demoUser.Role)
	if err := db.Save(demoUser).Error; err != nil {
		return errors.Wrap(err, "failed to migrate demo user to DEMO role")
	}
	return nil
}

func migrateNewPermissions(db *gorm.DB) ([]*models.PermissionModel, error) {
	newPermissions := make([]*models.PermissionModel, 0)
	for _, perm := range models.AllPermission {
		var existingPermission *models.PermissionModel
		if err := db.Where("name = ?", perm).First(&existingPermission).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			newPermissions = append(newPermissions, &models.PermissionModel{Name: perm})
		}
	}
	if len(newPermissions) == 0 {
		return make([]*models.PermissionModel, 0), nil
	}

	return newPermissions, db.Save(&newPermissions).Error
}
