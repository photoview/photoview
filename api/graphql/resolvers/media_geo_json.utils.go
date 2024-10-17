package resolvers

type geoMedia struct {
	MediaID         int
	MediaTitle      string
	ThumbnailName   string
	ThumbnailWidth  int
	ThumbnailHeight int
	Latitude        float64
	Longitude       float64
}

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
