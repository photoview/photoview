package resolvers

import (
	"errors"
	"net/url"
)

// validateMapStyleURL checks that a map style URL is valid and uses http or https.
func validateMapStyleURL(rawURL string) error {
	if len(rawURL) > 2048 {
		return errors.New("map style URL must not exceed 2048 characters")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("invalid map style URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("map style URL scheme must be http or https")
	}

	if parsed.Host == "" {
		return errors.New("map style URL must have a host")
	}

	return nil
}
