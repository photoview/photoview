package models

import (
	"fmt"
	"time"
)

type MediaEXIF struct {
	Model
	Description     *string
	Camera          *string
	Maker           *string
	Lens            *string
	DateShot        *time.Time
	OffsetSecShot   *int
	Exposure        *float64
	Aperture        *float64
	Iso             *int64
	FocalLength     *float64
	Flash           *int64
	Orientation     *int64
	ExposureProgram *int64
	GPSLatitude     *float64
	GPSLongitude    *float64
}

func (MediaEXIF) TableName() string {
	return "media_exif"
}

func (exif *MediaEXIF) Media() *Media {
	panic("not implemented")
}

func (exif *MediaEXIF) Coordinates() *Coordinates {
	if exif.GPSLatitude == nil || exif.GPSLongitude == nil {
		return nil
	}

	return &Coordinates{
		Latitude:  *exif.GPSLatitude,
		Longitude: *exif.GPSLongitude,
	}
}

func (exif *MediaEXIF) DateShotWithOffset() time.Time {
	if exif.DateShot == nil {
		return time.Time{}
	}

	if exif.OffsetSecShot == nil {
		return *exif.DateShot
	}

	offsetAbs := *exif.OffsetSecShot
	sign := "+"
	if offsetAbs < 0 {
		offsetAbs = -offsetAbs
		sign = "-"
	}
	hour := offsetAbs / 60 / 60
	minute := offsetAbs / 60 % 60
	zone := fmt.Sprintf("%s%02d:%02d", sign, hour, minute)

	loc := time.FixedZone(zone, *exif.OffsetSecShot)
	return exif.DateShot.In(loc)
}
