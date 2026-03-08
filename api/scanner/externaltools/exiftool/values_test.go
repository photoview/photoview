package exiftool

import (
	"math"
	"testing"
	"time"
)

func TestGPSIsValid(t *testing.T) {
	tests := []struct {
		name      string
		gps       GPS
		wantValid bool
	}{
		{"LatNormalLongNormal", GPS{10.0, 10.0, "10 10"}, true},
		{"ZeroValue", GPS{0, 0, "0 0"}, true},

		{"LatNilLongNormal", GPS{math.NaN(), 10.0, "NaN 10"}, false},
		{"LatNormalLongNil", GPS{10.0, math.NaN(), "10 NaN"}, false},

		{"Lat>90LongNormal", GPS{100.0, 10.0, "100 10"}, false},
		{"Lat<-90LongNormal", GPS{-100.0, 10.0, "-100 10"}, false},

		{"LatNormalLong>180", GPS{10.0, 190.0, "10 190"}, false},
		{"LatNormalLong<-180", GPS{10.0, -190.0, "10 -190"}, false},

		{"Empty", GPS{0, 0, ""}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.gps.IsValid()
			if got != tc.wantValid {
				t.Fatalf("gps.IsValid(%v) = %v, want: %v", tc.gps, got, tc.wantValid)
			}
		})
	}
}

func mustParseInLocation(t *testing.T, timeStr string) time.Time {
	t.Helper()

	ret, err := time.ParseInLocation(layout, timeStr, time.Local)
	if err != nil {
		t.Fatalf("time.ParseInLocation(%q) returns error: %v", timeStr, err)
	}

	return ret
}

func TestTimeAllTimeLocal(t *testing.T) {
	wantStr := "2025:10:28 14:20:22.164"
	otherStr := "2024:10:28 14:20:22.164"

	tests := []struct {
		name    string
		timeAll TimeAll
		want    time.Time
	}{
		{"DateTimeOriginal", TimeAll{
			DateTimeOriginal: wantStr,
			CreateDate:       otherStr,
			TrackCreateDate:  otherStr,
			MediaCreateDate:  otherStr,
			FileModifyDate:   otherStr,
		}, mustParseInLocation(t, wantStr)},
		{"CreateDate", TimeAll{
			DateTimeOriginal: "",
			CreateDate:       wantStr,
			TrackCreateDate:  otherStr,
			MediaCreateDate:  otherStr,
			FileModifyDate:   otherStr,
		}, mustParseInLocation(t, wantStr)},
		{"TrackCreateDate", TimeAll{
			DateTimeOriginal: "",
			CreateDate:       "",
			TrackCreateDate:  wantStr,
			MediaCreateDate:  otherStr,
			FileModifyDate:   otherStr,
		}, mustParseInLocation(t, wantStr)},
		{"MediaCreateDate", TimeAll{
			DateTimeOriginal: "",
			CreateDate:       "",
			TrackCreateDate:  "",
			MediaCreateDate:  wantStr,
			FileModifyDate:   otherStr,
		}, mustParseInLocation(t, wantStr)},
		{"FileModifyDate", TimeAll{
			DateTimeOriginal: "",
			CreateDate:       "",
			TrackCreateDate:  "",
			MediaCreateDate:  "",
			FileModifyDate:   wantStr,
		}, mustParseInLocation(t, wantStr)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, isLocal := tc.timeAll.Time()
			if !isLocal {
				t.Fatalf("timeAll.Time() is not a local time, which it should be")
			}

			if !got.Equal(tc.want) {
				t.Errorf("timeAll.Time() = %v, want: %v", got, tc.want)
			}
		})
	}
}

func mustParse(t *testing.T, timeStr string) time.Time {
	t.Helper()

	ret, err := time.Parse(layoutWithOffset, timeStr)
	if err != nil {
		t.Fatalf("time.Parse(%q) returns error: %v", timeStr, err)
	}

	return ret
}

func TestTimeAllTimeWithTimezone(t *testing.T) {
	wantStr := "2025:10:28 14:20:22.164+01:00"
	otherStr := "2024:10:28 14:20:22.164+01:00"

	tests := []struct {
		name    string
		timeAll TimeAll
		want    time.Time
	}{
		{"SubSecDateTimeOriginal", TimeAll{
			SubSecDateTimeOriginal: wantStr,
			SubSecCreateDate:       otherStr,
			DateTimeOriginal:       otherStr,
			CreateDate:             otherStr,
			TrackCreateDate:        otherStr,
			MediaCreateDate:        otherStr,
			FileModifyDate:         otherStr,
		}, mustParse(t, wantStr)},
		{"SubSecCreateDate", TimeAll{
			SubSecDateTimeOriginal: "",
			SubSecCreateDate:       wantStr,
			DateTimeOriginal:       otherStr,
			CreateDate:             otherStr,
			TrackCreateDate:        otherStr,
			MediaCreateDate:        otherStr,
			FileModifyDate:         otherStr,
		}, mustParse(t, wantStr)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, isLocal := tc.timeAll.Time()
			if isLocal {
				t.Fatalf("timeAll.Time() is a local time, which it should not be")
			}

			if !got.Equal(tc.want) {
				t.Errorf("timeAll.Time() = %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestTimeAllTimeEmpty(t *testing.T) {
	var timeAll TimeAll
	got, isLocal := timeAll.Time()
	if isLocal {
		t.Fatalf("timeAll.Time() is a local time, which it should not be")
	}

	if !got.IsZero() {
		t.Errorf("timeAll.Time() is a valid time, which it should not be")
	}
}

func TestTimeAllOffsetSecs(t *testing.T) {
	wantStr := "+01:00"
	otherStr := "-01:00"
	gpsDate := "2025:10:28"
	gpsTime := "14:20:22"
	localTime := mustParse(t, gpsDate+" "+gpsTime+otherStr)

	tests := []struct {
		name    string
		timeAll TimeAll
		want    int
	}{
		{"OffsetTimeOriginal", TimeAll{
			OffsetTimeOriginal: wantStr,
			OffsetTime:         otherStr,
			TimeZone:           otherStr,
			GPSTimeStamp:       gpsTime,
			GPSDateStamp:       gpsDate,
		}, 60 * 60},
		{"OffsetTime", TimeAll{
			OffsetTimeOriginal: "",
			OffsetTime:         wantStr,
			TimeZone:           otherStr,
			GPSTimeStamp:       gpsTime,
			GPSDateStamp:       gpsDate,
		}, 60 * 60},
		{"TimeZone", TimeAll{
			OffsetTimeOriginal: "",
			OffsetTime:         "",
			TimeZone:           wantStr,
			GPSTimeStamp:       gpsTime,
			GPSDateStamp:       gpsDate,
		}, 60 * 60},
		{"GPS", TimeAll{
			OffsetTimeOriginal: "",
			OffsetTime:         "",
			TimeZone:           "",
			GPSTimeStamp:       gpsTime,
			GPSDateStamp:       gpsDate,
		}, -60 * 60},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.timeAll.OffsetSecs(localTime)
			if !ok {
				t.Fatalf("timeAll.OffsetSecs() is not valid, which it should be")
			}

			if got != tc.want {
				t.Errorf("timeAll.OffsetSecs() = %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestTimeAllOffsetSecsEmptyLocal(t *testing.T) {
	timeAll := TimeAll{
		OffsetTimeOriginal: "",
		OffsetTime:         "",
		TimeZone:           "",
		GPSTimeStamp:       "2025:10:28",
		GPSDateStamp:       "14:20:22",
	}

	got, ok := timeAll.OffsetSecs(time.Time{})
	if ok || got != 0 {
		t.Errorf("timeAll.OffsetSecs() = (%d, %v), want: (%d, %v)", got, ok, 0, false)
	}
}
