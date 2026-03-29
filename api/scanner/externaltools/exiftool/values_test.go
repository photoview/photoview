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
		{"LatNormalLongNormal", GPS{new(10.0), new(10.0)}, true},
		{"ZeroValue", GPS{new(0.0), new(0.0)}, true},

		{"LatNilLongNormal", GPS{new(math.NaN()), new(10.0)}, false},
		{"LatNormalLongNil", GPS{new(10.0), new(math.NaN())}, false},

		{"Lat>90LongNormal", GPS{new(100.0), new(10.0)}, false},
		{"Lat<-90LongNormal", GPS{new(-100.0), new(10.0)}, false},

		{"LatNormalLong>180", GPS{new(10.0), new(190.0)}, false},
		{"LatNormalLong<-180", GPS{new(10.0), new(-190.0)}, false},

		{"Empty", GPS{}, false},
		{"EmptyLat", GPS{nil, new(0.0)}, false},
		{"EmptyLong", GPS{new(0.0), nil}, false},
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

func mustParseInUTC(t *testing.T, timeStr string) time.Time {
	t.Helper()

	ret, err := time.ParseInLocation(layout, timeStr, time.UTC)
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
			DateTimeOriginal: &wantStr,
			CreateDate:       &otherStr,
			TrackCreateDate:  &otherStr,
			MediaCreateDate:  &otherStr,
			FileModifyDate:   &otherStr,
		}, mustParseInUTC(t, wantStr)},
		{"CreateDate", TimeAll{
			CreateDate:      &wantStr,
			TrackCreateDate: &otherStr,
			MediaCreateDate: &otherStr,
			FileModifyDate:  &otherStr,
		}, mustParseInUTC(t, wantStr)},
		{"TrackCreateDate", TimeAll{
			TrackCreateDate: &wantStr,
			MediaCreateDate: &otherStr,
			FileModifyDate:  &otherStr,
		}, mustParseInUTC(t, wantStr)},
		{"MediaCreateDate", TimeAll{
			MediaCreateDate: &wantStr,
			FileModifyDate:  &otherStr,
		}, mustParseInUTC(t, wantStr)},
		{"FileModifyDate", TimeAll{
			FileModifyDate: &wantStr,
		}, mustParseInUTC(t, wantStr)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.timeAll.TimeInLocal(); !got.Equal(tc.want) {
				t.Errorf("timeAll.Time() = %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestTimeAllTimeWithTimezone(t *testing.T) {
	wantStr := "2025:10:28 14:20:22.164"
	otherStr := "2024:10:28 14:20:22.164"

	tests := []struct {
		name    string
		timeAll TimeAll
		want    time.Time
	}{
		{"SubSecDateTimeOriginal", TimeAll{
			SubSecDateTimeOriginal: &wantStr,
			SubSecCreateDate:       &otherStr,
			DateTimeOriginal:       &otherStr,
			CreateDate:             &otherStr,
			TrackCreateDate:        &otherStr,
			MediaCreateDate:        &otherStr,
			FileModifyDate:         &otherStr,
		}, mustParseInUTC(t, wantStr)},
		{"SubSecCreateDate", TimeAll{
			SubSecCreateDate: &wantStr,
			DateTimeOriginal: &otherStr,
			CreateDate:       &otherStr,
			TrackCreateDate:  &otherStr,
			MediaCreateDate:  &otherStr,
			FileModifyDate:   &otherStr,
		}, mustParseInUTC(t, wantStr)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.timeAll.TimeInLocal(); !got.Equal(tc.want) {
				t.Errorf("timeAll.Time() = %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestTimeAllTimeEmpty(t *testing.T) {
	var timeAll TimeAll
	if got := timeAll.TimeInLocal(); !got.IsZero() {
		t.Errorf("timeAll.Time() is a valid time, which it should not be")
	}
}

func TestTimeAllOffsetSecs(t *testing.T) {
	wantStr := "+01:00"
	otherStr := "-01:00"
	gpsDateTime := "2025:10:28 14:20:22Z"
	otherTimezone := -120
	localTime := mustParseInUTC(t, "2025:10:28 13:20:22")

	tests := []struct {
		name    string
		timeAll TimeAll
		want    int
	}{
		{"OffsetTimeOriginal", TimeAll{
			OffsetTimeOriginal: &wantStr,
			OffsetTime:         &otherStr,
			TimeZone:           &otherTimezone,
			GPSDateTime:        &gpsDateTime,
		}, 60 * 60},
		{"OffsetTime", TimeAll{
			OffsetTime:  &wantStr,
			TimeZone:    &otherTimezone,
			GPSDateTime: &gpsDateTime,
		}, 60 * 60},
		{"TimeZone", TimeAll{
			TimeZone:    &otherTimezone,
			GPSDateTime: &gpsDateTime,
		}, -120 * 60},
		{"GPS", TimeAll{
			GPSDateTime: &gpsDateTime,
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
		GPSDateTime: new("14:20:22 2025:10:28Z"),
	}

	got, ok := timeAll.OffsetSecs(time.Time{})
	if ok || got != 0 {
		t.Errorf("timeAll.OffsetSecs() = (%d, %v), want: (%d, %v)", got, ok, 0, false)
	}
}
