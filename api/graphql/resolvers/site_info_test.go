package resolvers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMapStyleURL(t *testing.T) {
	t.Run("valid HTTPS URL", func(t *testing.T) {
		err := validateMapStyleURL("https://tiles.openfreemap.org/styles/liberty")
		assert.NoError(t, err)
	})

	t.Run("valid HTTP URL", func(t *testing.T) {
		err := validateMapStyleURL("http://tiles.example.com/style.json")
		assert.NoError(t, err)
	})

	t.Run("javascript scheme rejected", func(t *testing.T) {
		err := validateMapStyleURL("javascript:alert(1)")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scheme must be http or https")
	})

	t.Run("data scheme rejected", func(t *testing.T) {
		err := validateMapStyleURL("data:text/html,<script>alert(1)</script>")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scheme must be http or https")
	})

	t.Run("file scheme rejected", func(t *testing.T) {
		err := validateMapStyleURL("file:///etc/passwd")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scheme must be http or https")
	})

	t.Run("empty string rejected", func(t *testing.T) {
		err := validateMapStyleURL("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scheme must be http or https")
	})

	t.Run("overly long URL rejected", func(t *testing.T) {
		longURL := "https://example.com/" + strings.Repeat("a", 2048)
		err := validateMapStyleURL(longURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must not exceed 2048 characters")
	})

	t.Run("missing host rejected", func(t *testing.T) {
		err := validateMapStyleURL("https://")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have a host")
	})
}
