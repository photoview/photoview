package exiftool

import (
	"fmt"
	"math"
	"time"
)

// GPS stores gps-related tags.
type GPS struct {
	Latitude  float64
	Longitude float64
}

// IsValid returns true when GPS data is valid.
func (gps GPS) IsValid() bool {
	if math.IsNaN(gps.Latitude) {
		return false
	}

	if math.IsNaN(gps.Longitude) {
		return false
	}

	if math.Abs(gps.Latitude) > 90 || math.Abs(gps.Longitude) > 180 {
		return false
	}

	return true
}

func (gps GPS) String() string {
	return fmt.Sprintf("GPS(%.9f, %.9f)", gps.Latitude, gps.Longitude)
}

// TimeAll stores tags returned by -time:all.
type TimeAll struct {
	SubSecDateTimeOriginal string
	SubSecCreateDate       string

	DateTimeOriginal string
	CreateDate       string
	TrackCreateDate  string
	MediaCreateDate  string
	FileModifyDate   string

	OffsetTimeOriginal string
	OffsetTime         string
	TimeZone           string

	GPSDateStamp string
	GPSTimeStamp string
}

const layout = "2006:01:02 15:04:05.999"
const layoutWithOffset = "2006:01:02 15:04:05.999-07:00"

// Time returns most likely time. True returns if the time is a local time without a timezone.
func (t TimeAll) Time() (time.Time, bool) {
	for _, dateStr := range []string{
		// Keep the order for the priority to generate DateShot
		t.SubSecDateTimeOriginal,
		t.SubSecCreateDate,
		t.DateTimeOriginal,
		t.CreateDate,
		t.TrackCreateDate,
		t.MediaCreateDate,
		t.FileModifyDate,
	} {
		if date, err := time.ParseInLocation(layout, dateStr, time.Local); err == nil {
			return date, true
		}

		if date, err := time.Parse(layoutWithOffset, dateStr); err == nil {
			return date, false
		}
	}

	return time.Time{}, false
}

// OffsetSecs returns seconds offset by UTC.
func (t TimeAll) OffsetSecs(local time.Time) (int, bool) {
	for _, str := range []string{
		t.OffsetTimeOriginal,
		t.OffsetTime,
		t.TimeZone,
	} {
		if t, err := time.Parse("-07:00", str); err == nil {
			_, offsetSecs := t.Zone()
			return offsetSecs, true
		}
	}

	if local.IsZero() {
		return 0, false
	}

	gpsDate, err := time.ParseInLocation(layout, t.GPSDateStamp+" "+t.GPSTimeStamp, time.UTC)
	if err != nil {
		return 0, false
	}

	// GPS time is always UTC per EXIF spec
	// offset = GPS UTC time - local time
	offset := int(gpsDate.Sub(local).Seconds())
	return offset, true
}
