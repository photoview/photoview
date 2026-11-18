package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeUserPreferencesSearchResultLimit(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "preferences_user", nil, false)
	assert.NoError(t, err)

	r := &mutationResolver{Resolver: &Resolver{database: db}}
	ctx := auth.AddUserToContext(context.Background(), user)

	t.Run("an unset limit leaves the server default in place", func(t *testing.T) {
		prefs, err := r.ChangeUserPreferences(ctx, nil, nil)
		assert.NoError(t, err)
		if assert.NotNil(t, prefs) {
			assert.Nil(t, prefs.SearchResultLimit)
		}
	})

	t.Run("a limit is stored and read back", func(t *testing.T) {
		limit := 25
		prefs, err := r.ChangeUserPreferences(ctx, nil, &limit)
		assert.NoError(t, err)
		if assert.NotNil(t, prefs.SearchResultLimit) {
			assert.Equal(t, 25, *prefs.SearchResultLimit)
		}

		var stored models.UserPreferences
		assert.NoError(t, db.Where("user_id = ?", user.ID).First(&stored).Error)
		if assert.NotNil(t, stored.SearchResultLimit) {
			assert.Equal(t, 25, *stored.SearchResultLimit)
		}
	})

	t.Run("zero is kept as a value, not treated as unset", func(t *testing.T) {
		unlimited := 0
		prefs, err := r.ChangeUserPreferences(ctx, nil, &unlimited)
		assert.NoError(t, err)
		if assert.NotNil(t, prefs.SearchResultLimit, "0 means unlimited and must survive the round trip") {
			assert.Equal(t, 0, *prefs.SearchResultLimit)
		}
	})

	t.Run("a negative value clears the preference", func(t *testing.T) {
		limit := 25
		_, err := r.ChangeUserPreferences(ctx, nil, &limit)
		assert.NoError(t, err)

		// 0 already means unlimited, so it cannot also mean "forget my
		// setting" - without this there would be no way back to the default.
		clear := -1
		prefs, err := r.ChangeUserPreferences(ctx, nil, &clear)
		assert.NoError(t, err)
		if assert.NotNil(t, prefs) {
			assert.Nil(t, prefs.SearchResultLimit)
		}

		var stored models.UserPreferences
		assert.NoError(t, db.Where("user_id = ?", user.ID).First(&stored).Error)
		assert.Nil(t, stored.SearchResultLimit)
	})

	t.Run("changing one preference leaves the others alone", func(t *testing.T) {
		limit := 40
		_, err := r.ChangeUserPreferences(ctx, nil, &limit)
		assert.NoError(t, err)

		language := string(models.LanguageTranslationGerman)
		prefs, err := r.ChangeUserPreferences(ctx, &language, nil)
		assert.NoError(t, err)
		if !assert.NotNil(t, prefs) {
			return
		}

		if assert.NotNil(t, prefs.SearchResultLimit, "a language change must not reset the search limit") {
			assert.Equal(t, 40, *prefs.SearchResultLimit)
		}
		if assert.NotNil(t, prefs.Language) {
			assert.Equal(t, models.LanguageTranslationGerman, *prefs.Language)
		}
	})
}
