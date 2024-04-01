package auth

import "github.com/photoview/photoview/api/graphql/models"

// This section defines the default user roles available during migration and creation of the system.
var (
	// USER defines the default permissions available for the USER role.
	USER = []models.Permission{}
	// DEMO defines the default permissions available for the DEMO role.
	DEMO = []models.Permission{}
)
