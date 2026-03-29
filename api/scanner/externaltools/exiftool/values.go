package exiftool

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// GPS stores gps-related tags.
type GPS struct {
	GPSLatitude  *float64
	GPSLongitude *float64
}

// IsValid returns true when GPS data is valid.
func (gps GPS) IsValid() bool {
	if gps.GPSLongitude == nil || gps.GPSLatitude == nil {
		return false
	}

	if math.IsNaN(*gps.GPSLatitude) {
		return false
	}

	if math.IsNaN(*gps.GPSLongitude) {
		return false
	}

	if math.Abs(*gps.GPSLatitude) > 90 || math.Abs(*gps.GPSLongitude) > 180 {
		return false
	}

	return true
}

func (gps GPS) String() string {
	if !gps.IsValid() {
		return "GPS(invalid)"
	}

	return fmt.Sprintf("GPS(%.9f, %.9f)", *gps.GPSLatitude, *gps.GPSLongitude)
}

// TimeAll stores tags returned by -time:all.
type TimeAll struct {
	SubSecDateTimeOriginal *string
	SubSecCreateDate       *string

	DateTimeOriginal *string
	CreateDate       *string
	TrackCreateDate  *string
	MediaCreateDate  *string
	FileModifyDate   *string

	OffsetTimeOriginal *string
	OffsetTime         *string
	TimeZone           *int

	GPSDateTime *string
}

const layout = "2006:01:02 15:04:05.999"
const layoutWithTimezone = "2006:01:02 15:04:05.999Z07:00"

// TimeInLocal returns most likely time. The date and time are in local. The timezone is meaningless and always be in UTC. Use `OffsetSecs()` to determine the timezone.
func (t TimeAll) TimeInLocal() time.Time {
	for _, dateP := range []*string{
		// Keep the order for the priority to generate DateShot
		t.SubSecDateTimeOriginal,
		t.SubSecCreateDate,
		t.DateTimeOriginal,
		t.CreateDate,
		t.TrackCreateDate,
		t.MediaCreateDate,
		t.FileModifyDate,
	} {
		if dateP == nil {
			continue
		}

		date := *dateP

		// Ignore timezone
		if zoneIndex := strings.IndexAny(date, "+-Z"); zoneIndex >= 0 {
			date = date[:zoneIndex]
		}

		if date, err := time.ParseInLocation(layout, date, time.UTC); err == nil {
			return date
		}
	}

	return time.Time{}
}

// OffsetSecs returns seconds offset by UTC.
func (t TimeAll) OffsetSecs(local time.Time) (int, bool) {
	for _, offsetP := range []*string{
		t.OffsetTimeOriginal,
		t.OffsetTime,
	} {
		if offsetP == nil {
			continue
		}

		if t, err := time.Parse("-07:00", *offsetP); err == nil {
			_, offsetSecs := t.Zone()
			return offsetSecs, true
		}
	}

	// TimeZone is in minutes
	if t.TimeZone != nil {
		return *t.TimeZone * 60, true
	}

	// Calculate offset sec with GPS time and local time.
	if local.IsZero() {
		return 0, false
	}

	if t.GPSDateTime == nil {
		return 0, false
	}

	gpsDate, err := time.Parse(layoutWithTimezone, *t.GPSDateTime)
	if err != nil {
		return 0, false
	}
	gpsDate = gpsDate.UTC()

	// GPS time is always UTC per EXIF spec
	// offset = local time (in UTC) - GPS UTC time
	offset := int(local.Sub(gpsDate).Seconds())
	return offset, true
}

type PhotoMeta struct {
	ImageDescription *string
	Model            *string
	Make             *string
	LensModel        *string
	ISO              *int64
	Flash            *int64
	Orientation      *int64
	ExposureProgram  *int64
	ExposureTime     *float64
	Aperture         *float64
	FocalLength      *float64
}

func (m *PhotoMeta) SanitizeFloats() {
	for _, value := range []**float64{
		&m.ExposureTime,
		&m.Aperture,
		&m.FocalLength,
	} {
		if *value == nil {
			continue
		}

		if math.IsNaN(**value) || math.IsInf(**value, 0) {
			*value = nil
		}
	}
}

type MIMEType struct {
	MIMEType *string
}
