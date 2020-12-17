package resolvers

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/utils"
)

type geoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties interface{}            `json:"properties"`
	Geometry   geoJSONFeatureGeometry `json:"geometry"`
}

type geoJSONMediaProperties struct {
	MediaID    int    `json:"media_id"`
	MediaTitle string `json:"media_title"`
	Thumbnail  struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail"`
}

type geoJSONFeatureGeometry struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

func makeGeoJSONFeatureCollection(features []geoJSONFeature) geoJSONFeatureCollection {
	return geoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

func makeGeoJSONFeature(properties interface{}, geometry geoJSONFeatureGeometry) geoJSONFeature {
	return geoJSONFeature{
		Type:       "Feature",
		Properties: properties,
		Geometry:   geometry,
	}
}

func makeGeoJSONFeatureGeometryPoint(lat float64, long float64) geoJSONFeatureGeometry {
	coordinates := [2]float64{long, lat}

	return geoJSONFeatureGeometry{
		Type:        "Point",
		Coordinates: coordinates,
	}
}

func (r *queryResolver) MyMediaGeoJSON(ctx context.Context) (interface{}, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	rows, err := r.Database.Query(`
	SELECT media.media_id, media.title,
	  url.media_name AS thumbnail_name, url.width AS thumbnail_width, url.height AS thumbnail_height,
	  exif.gps_latitude, exif.gps_longitude FROM media_exif exif
	INNER JOIN media ON exif.exif_id = media.exif_id
	INNER JOIN media_url url ON media.media_id = url.media_id
	INNER JOIN album ON media.album_id = album.album_id
	WHERE exif.gps_latitude IS NOT NULL
	  AND exif.gps_longitude IS NOT NULL
		AND url.purpose = 'thumbnail'
		AND album.owner_id = ?;
	`, user.UserID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	features := make([]geoJSONFeature, 0)

	for rows.Next() {

		var mediaID int
		var mediaTitle string
		var thumbnailName string
		var thumbnailWidth int
		var thumbnailHeight int
		var latitude float64
		var longitude float64

		if err := rows.Scan(&mediaID, &mediaTitle, &thumbnailName, &thumbnailWidth, &thumbnailHeight, &latitude, &longitude); err != nil {
			return nil, err
		}

		geoPoint := makeGeoJSONFeatureGeometryPoint(latitude, longitude)

		thumbnailURL := utils.ApiEndpointUrl()
		thumbnailURL.Path = path.Join(thumbnailURL.Path, "photo", thumbnailName)

		properties := geoJSONMediaProperties{
			MediaID:    mediaID,
			MediaTitle: mediaTitle,
			Thumbnail: struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			}{
				URL:    thumbnailURL.String(),
				Width:  thumbnailWidth,
				Height: thumbnailHeight,
			},
		}

		features = append(features, makeGeoJSONFeature(properties, geoPoint))
	}

	featureCollection := makeGeoJSONFeatureCollection(features)

	return featureCollection, nil
}

func (r *queryResolver) MapboxToken(ctx context.Context) (*string, error) {
	mapboxTokenEnv := os.Getenv("MAPBOX_TOKEN")
	if mapboxTokenEnv == "" {
		return nil, nil
	}

	return &mapboxTokenEnv, nil
}
