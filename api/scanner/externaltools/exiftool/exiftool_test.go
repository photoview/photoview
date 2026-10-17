package exiftool

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

func TestExiftool(t *testing.T) {
	instance, err := New()
	if err != nil {
		t.Fatalf("create instance error: %v", err)
	}

	t.Log("bin:", instance.BinaryPath())
	t.Log("version:", instance.Version())

	if instance.BinaryPath() == "" {
		t.Errorf("want exiftool binary, but got an empty string")
	}

	if instance.Version() == "" {
		t.Errorf("want exiftool version, but got an emtpy string")
	}

	if err := instance.Close(); err != nil {
		t.Errorf("close instance error: %v", err)
	}
}

func TestExiftoolQueryJSONTagsWithEmbed(t *testing.T) {
	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	file := "./test_data/correct_gps.jpg"
	var value struct {
		TimeAll
		MIMEType
		PhotoMeta
	}

	if err := instance.QueryJSONTagsByNumber(file, &value); err != nil {
		t.Fatalf("QueryJSONTagsByNumber(%q) error: %v", file, err)
		return
	}

	if time := value.TimeAll.TimeInLocal(); time.IsZero() {
		t.Errorf("QueryJSONTagsByNumber(%q) error: no valid TimeAll", file)
	}
	if value.MIMEType.MIMEType == nil {
		t.Errorf("QueryJSONTagsByNumber(%q) error: no valid MIMEType", file)
	}
	if value.PhotoMeta.Model == nil {
		t.Errorf("QueryJSONTagsByNumber(%q) error: no valid PhotoMeta", file)
	}
}

func TestExiftoolQueryMIMEType(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"./test_data/no_timezone.jpg", "image/jpeg"},
		{"./test_data/subsec_timezone.heic", "image/heic"},
		{"./test_data/raw_with_preview_jpg.cr3", "image/x-canon-cr3"},
		{"./test_data/no_exif.jpg", "image/jpeg"},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			var value struct{ MIMEType }
			err := instance.QueryJSONTagsByNumber(tc.file, &value)
			if err != nil {
				t.Errorf("QueryJSONTagsByNumber(%q) error: %v", tc.file, err)
				return
			}

			if got := value.MIMEType.MIMEType; got == nil || *got != tc.want {
				t.Errorf("QueryJSONTagsByNumber(%q) = %v, want: %q", tc.file, got, tc.want)
			}
		})
	}
}

func checkTimeallFieldsHasValue(t *testing.T, time TimeAll, fields []string) {
	t.Helper()

	tv := reflect.ValueOf(time)

	for _, field := range fields {
		fv := tv.FieldByName(field)
		if !fv.IsValid() {
			t.Errorf("can't find field %q in TimeAll", field)
			continue
		}

		if fv.IsNil() || fv.Elem().IsZero() {
			t.Errorf("field %q is type %T with value %v", field, fv.Type(), fv.Interface())
			continue
		}
	}
}

func TestExiftoolQueryTimeAllHasOffset(t *testing.T) {
	tests := []struct {
		file          string
		wantKeys      []string
		wantTime      time.Time
		wantOffsetSec int
	}{
		{"./test_data/raw_with_preview_jpg.cr3", []string{
			"CreateDate",
			"DateTimeOriginal",
			"FileModifyDate",
			"MediaCreateDate",
			"OffsetTime",
			"OffsetTimeOriginal",
			"SubSecCreateDate",
			"SubSecDateTimeOriginal",
			"TimeZone",
			"TrackCreateDate",
		}, mustParseInUTC(t, "2019:09:13 14:36:48.87"), 7200},
		{"./test_data/subsec_timezone.heic", []string{
			"CreateDate",
			"DateTimeOriginal",
			"FileModifyDate",
			"OffsetTime",
			"OffsetTimeOriginal",
			"SubSecCreateDate",
			"SubSecDateTimeOriginal",
		}, mustParseInUTC(t, "2025:10:28 14:20:22.164"), 3600},
		{"./test_data/createdate_timezone_separate.jpg", []string{
			"CreateDate",
			"DateTimeOriginal",
			"TimeZone",
			"FileModifyDate",
			"SubSecCreateDate",
			"SubSecDateTimeOriginal",
		}, mustParseInUTC(t, "2008:03:15 07:44:21.49"), -7 * 60 * 60},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			var value struct{ TimeAll }
			err := instance.QueryJSONTagsByNumber(tc.file, &value)
			if err != nil {
				t.Errorf("QueryJSONTagsByNumber(%q) error: %v", tc.file, err)
				return
			}

			checkTimeallFieldsHasValue(t, value.TimeAll, tc.wantKeys)

			gotTime := value.TimeAll.TimeInLocal()
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("value.TimeAll.TimeInLocal() = %v, want: %v", gotTime, tc.wantTime)
			}

			if got, ok := value.TimeAll.OffsetSecs(gotTime); !ok || got != tc.wantOffsetSec {
				t.Errorf("value.TimeAll.OffsetSecs() = (%v, %v), want: (%v, true)", got, ok, tc.wantOffsetSec)
			}
		})
	}
}

const (
	fakeExiftoolControlEnv = "EXIFTOOL_FAKE_CONTROL"
	stuckQueryTimeout      = 500 * time.Millisecond
	queryReturnBound       = 5 * time.Second
)

const fakeExiftoolScript = `#!/usr/bin/perl
use strict;
use warnings;
$| = 1;
{
    my $old = select STDERR;
    $| = 1;
    select $old;
}

if (@ARGV >= 1 && $ARGV[0] eq '-ver') {
    print "0.0.0-fake\n";
    exit 0;
}

if (defined $ENV{EXIFTOOL_FAKE_PID} && $ENV{EXIFTOOL_FAKE_PID} ne '') {
    open my $fh, '>', $ENV{EXIFTOOL_FAKE_PID} or die $!;
    print {$fh} "$$\n";
    close $fh;
}

while (defined(my $line = <STDIN>)) {
    next if $line ne "-execute\n";

    my $mode = 'ok';
    if (defined $ENV{EXIFTOOL_FAKE_CONTROL} && open my $fh, '<', $ENV{EXIFTOOL_FAKE_CONTROL}) {
        if (defined(my $m = <$fh>)) {
            chomp $m;
            $mode = $m if $m ne '';
        }
        close $fh;
    }

    if ($mode eq 'hang') {
        sleep 3600 while 1;
    }

    print STDOUT qq([{"MIMEType":"image/jpeg","Model":"FakeCam"}]\n);
    print STDOUT "{ready}\n";
    print STDERR "{ready}\n";
}
`

func installFakeExiftool(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "exiftool")
	if err := os.WriteFile(path, []byte(fakeExiftoolScript), 0o755); err != nil {
		t.Fatalf("write fake exiftool: %v", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("chmod fake exiftool: %v", err)
	}

	return dir
}

func fileModifyDateLiteralUTC(t *testing.T, file string) time.Time {
	t.Helper()

	dir := filepath.Dir(file)
	file = filepath.Base(file)

	fsys := os.DirFS(dir)
	info, err := fs.Stat(fsys, file)
	if err != nil {
		t.Fatalf("read file stat error: %v", err)
	}

	ret := info.ModTime().Truncate(time.Second)

	_, offsetSec := ret.Zone()
	ret = ret.Add(time.Duration(offsetSec) * time.Second).UTC()

	return ret
}

func TestExiftoolQueryTimeAllNoOffset(t *testing.T) {
	tests := []struct {
		file          string
		wantKeys      []string
		wantTime      time.Time
		wantHasOffset bool
		wantOffset    int
	}{
		{"./test_data/no_timezone.jpg", []string{
			"DateTimeOriginal",
			"FileModifyDate",
		}, mustParseInUTC(t, "2012:05:06 15:39:44"), false, 0},
		{"./test_data/subsec_no_timezone.heic", []string{
			"CreateDate",
			"DateTimeOriginal",
			"FileModifyDate",
			"OffsetTime",
			"SubSecCreateDate",
			"SubSecDateTimeOriginal",
		}, mustParseInUTC(t, "2025:10:28 14:20:22.164"), true, 3600},
		{"./test_data/no_exif.jpg", []string{
			"FileModifyDate",
		}, fileModifyDateLiteralUTC(t, "./test_data/no_exif.jpg"), false, 0},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			var value struct{ TimeAll }
			err := instance.QueryJSONTagsByNumber(tc.file, &value)
			if err != nil {
				t.Errorf("QueryJSONTagsByNumber(%q) error: %v", tc.file, err)
				return
			}

			checkTimeallFieldsHasValue(t, value.TimeAll, tc.wantKeys)

			gotTime := value.TimeAll.TimeInLocal()
			if !gotTime.Equal(tc.wantTime) {
				t.Errorf("value.TimeAll.TimeInLocal() = %v, want: %v", gotTime, tc.wantTime)
			}

			gotOffset, gotHasOffset := value.TimeAll.OffsetSecs(gotTime)
			if gotHasOffset != tc.wantHasOffset {
				t.Errorf("value.TimeAll.OffsetSecs() = (_, %v), want: (_, %v)", gotHasOffset, tc.wantHasOffset)
			}
			if !gotHasOffset {
				return
			}
			if gotOffset != tc.wantOffset {
				t.Errorf("value.TimeAll.OffsetSecs() = (%v, _), want: (%v, _)", gotOffset, tc.wantHasOffset)
			}
		})
	}
}

func TestExiftoolQueryGPS(t *testing.T) {
	tests := []struct {
		file              string
		hasGPS            bool
		wantLat, wantLong float64
	}{
		{"./test_data/correct_gps.jpg", true, 44.4789972, 11.2979222},
		{"./test_data/incorrect_gps.jpg", false, 0, 0},
		{"./test_data/no_exif.jpg", false, 0, 0},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			var value struct{ GPS }
			if err := instance.QueryJSONTagsByNumber(tc.file, &value); err != nil {
				t.Errorf("QueryJSONTagsByNumber(%q) error: %v", tc.file, err)
				return
			}

			hasGPS := value.GPS.IsValid()
			if hasGPS != tc.hasGPS {
				t.Errorf("QueryJSONTagsByNumber(%q) has GPS: %v, want GPS: %v", tc.file, hasGPS, tc.hasGPS)
				return
			}
			if !tc.hasGPS {
				return
			}

			gpsToString := func(latitude, longitude float64) string {
				return fmt.Sprintf("(%.7f, %.7f)", latitude, longitude)
			}

			lat := *value.GPS.GPSLatitude
			long := *value.GPS.GPSLongitude
			if got, want := gpsToString(lat, long), gpsToString(tc.wantLat, tc.wantLong); got != want {
				t.Errorf("QueryJSONTagsByNumber(%q) = %s, want: %s", tc.file, got, want)
			}
		})
	}
}

func TestExiftoolSaveJPEGPreview(t *testing.T) {
	tests := []struct {
		file   string
		wantOK bool
	}{
		{"./test_data/raw_with_preview_jpg.cr3", true},
		{"./test_data/no_timezone.jpg", false},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	outputDir := t.TempDir()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			output := filepath.Join(outputDir, "preview.jpg")
			ok, err := instance.SaveJPEGPreview(tc.file, output)
			if err != nil {
				t.Errorf("SaveJPEGPreview(%q, %q) error: %v", tc.file, output, err)
				return
			}

			if ok != tc.wantOK {
				t.Errorf("SaveJPEGPreview(%q, %q) = %v, want: %v", tc.file, output, ok, tc.wantOK)
			}

			if !ok {
				return
			}

			var jpg struct {
				MIMEType
				TimeAll
			}
			if err := instance.QueryJSONTagsByNumber(output, &jpg); err != nil {
				t.Fatalf("QueryJSONTagsByNumber(%q) error: %v", output, err)
				return
			}

			if got, want := jpg.MIMEType.MIMEType, "image/jpeg"; got == nil || *got != want {
				t.Errorf("MIMEType(%q) = %v, want: %q", output, got, want)
				return
			}

			var raw struct{ TimeAll }
			if err := instance.QueryJSONTagsByNumber(tc.file, &raw); err != nil {
				t.Fatalf("QueryJSONTagsByNumber(%q) error: %v", tc.file, err)
			}

			if got, want := jpg.TimeAll.TimeInLocal(), raw.TimeAll.TimeInLocal(); !got.Equal(want) {
				t.Errorf("jpg.TimeAll.TimeInLocal() = %q, want: %q", got, want)
			}
		})
	}

}

func TestExiftoolError(t *testing.T) {
	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	tests := []struct {
		file   string
		errStr string
	}{
		{"./test_data/non_exist.jpg", "File not found"},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			checkErr := func(err error, fmtStr string, args ...any) {
				t.Helper()

				if want := tc.errStr; err == nil || !strings.Contains(err.Error(), want) {
					t.Errorf(fmtStr+" %v, want %v", append(args, err, want)...)
				}
			}

			var value MIMEType
			err = instance.QueryJSONTagsByNumber(tc.file, &value)
			checkErr(err, "QueryJSONTagsByNumber(%q)", tc.file)

			output := filepath.Join(t.TempDir(), "output.jpg")
			_, err = instance.SaveJPEGPreview(tc.file, output)
			checkErr(err, "SaveJPEGPreview(%q, %q)", tc.file, output)
		})
	}

	var value MIMEType
	if err := instance.QueryJSONTagsByNumber("./test_data/no_exif.jpg", &value); err != nil {
		t.Fatalf("QueryJSONTagsByNumber after missing file error: %v", err)
	}
	if got := value.MIMEType; got == nil || *got != "image/jpeg" {
		t.Fatalf("QueryJSONTagsByNumber after missing file = %v, want image/jpeg", got)
	}
}

func TestQueryJSONTagsByNumberContext(t *testing.T) {
	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var missing MIMEType
	err = instance.QueryJSONTagsByNumberContext(ctx, "./test_data/non_exist.jpg", &missing)
	if err == nil || !strings.Contains(err.Error(), "File not found") {
		t.Fatalf("QueryJSONTagsByNumberContext(missing) = %v, want File not found", err)
	}

	var value MIMEType
	if err := instance.QueryJSONTagsByNumberContext(ctx, "./test_data/no_exif.jpg", &value); err != nil {
		t.Fatalf("QueryJSONTagsByNumberContext(%q) error: %v", "./test_data/no_exif.jpg", err)
	}
	if got := value.MIMEType; got == nil || *got != "image/jpeg" {
		t.Fatalf("QueryJSONTagsByNumberContext(%q) = %v, want image/jpeg", "./test_data/no_exif.jpg", got)
	}
}

func TestQueryJSONTagsByNumberContextDeadline(t *testing.T) {
	binDir := installFakeExiftool(t)
	control := filepath.Join(t.TempDir(), "control")
	if err := os.WriteFile(control, []byte("hang\n"), 0o644); err != nil {
		t.Fatalf("write control: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv(fakeExiftoolControlEnv, control)

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	pid := instance.cmd.Process.Pid
	t.Cleanup(func() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	})

	ctx, cancel := context.WithTimeout(context.Background(), stuckQueryTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		var value MIMEType
		done <- instance.QueryJSONTagsByNumberContext(ctx, "hang.jpg", &value)
	}()

	var queryErr error
	select {
	case queryErr = <-done:
	case <-time.After(queryReturnBound):
		_ = syscall.Kill(pid, syscall.SIGKILL)
		select {
		case <-done:
		case <-time.After(queryReturnBound):
		}
		t.Fatal("query blocked past the deadline")
	}

	if !errors.Is(queryErr, context.DeadlineExceeded) {
		t.Fatalf("query error = %v, want deadline exceeded", queryErr)
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- instance.Close()
	}()
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("Close() after timeout: %v", err)
		}
	case <-time.After(queryReturnBound):
		t.Fatal("Close() blocked; timed out process was not terminated")
	}

	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatalf("exiftool pid %d still running after timeout", pid)
	}
}
