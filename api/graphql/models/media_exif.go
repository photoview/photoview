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

const rfc3339WithoutTimezone = "2006-01-02T15:04:05.999"
const rfc3339WithTimezone = "2006-01-02T15:04:05.999-07:00"

func (exif *MediaEXIF) DateShotWithOffset() string {
	if exif.DateShot == nil {
		return ""
	}

	if exif.OffsetSecShot == nil {
		return exif.DateShot.Format(rfc3339WithoutTimezone)
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
	return exif.DateShot.In(loc).Format(rfc3339WithTimezone)
}
