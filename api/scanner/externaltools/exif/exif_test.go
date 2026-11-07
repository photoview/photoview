package exif

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

func resetForTest() {
	globalMu.Lock()
	defer globalMu.Unlock()
	if globalExifParser != nil {
		_ = globalExifParser.Close()
	}
	globalExifParser = nil
	// Allow Initialize() to run again in subsequent tests
	globalInit = sync.Once{}
}

func TestParseWithoutInit(t *testing.T) {
	resetForTest()
	if _, err := Parse("./test_data/bird.jpg"); err == nil {
		t.Fatalf("Parse() without Init() doesn't return an error")
	}
}

func TestParse(t *testing.T) {
	resetForTest()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	filename := "./test_data/bird.jpg"

	metadata, err := Parse(filename)
	if err != nil {
		t.Fatalf("Parse() returns an error: %v", err)
	}

	if metadata == nil {
		t.Errorf("Parse(%q) should not return nil", filename)
	}
}

func TestTimeoutRestartsParser(t *testing.T) {
	resetForTest()
	defer resetForTest()

	binDir := installFakeExiftool(t)
	control := filepath.Join(t.TempDir(), "control")
	pidPath := filepath.Join(t.TempDir(), "pid")
	if err := os.WriteFile(control, []byte("hang\n"), 0o644); err != nil {
		t.Fatalf("write control: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv(fakeExiftoolControlEnv, control)
	t.Setenv(fakeExiftoolPIDEnv, pidPath)

	prevTimeout := queryTimeout
	queryTimeout = stuckQueryTimeout
	defer func() { queryTimeout = prevTimeout }()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	t.Cleanup(func() {
		if pid := readPID(pidPath); pid > 0 {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})

	pid := waitPID(t, pidPath, 0)
	mustDeadline(t, pidPath, func() error {
		_, err := Parse("photo.jpg")
		return err
	})
	if !processGone(pid) {
		t.Fatalf("exiftool pid %d still running after Parse timeout", pid)
	}

	pid = waitPID(t, pidPath, pid)
	mustDeadline(t, pidPath, func() error {
		_, err := MIMEType("photo.jpg")
		return err
	})
	if !processGone(pid) {
		t.Fatalf("exiftool pid %d still running after MIMEType timeout", pid)
	}

	if err := os.WriteFile(control, []byte("ok\n"), 0o644); err != nil {
		t.Fatalf("write control: %v", err)
	}

	metadata, err := Parse("photo.jpg")
	if err != nil {
		t.Fatalf("Parse() after restart error: %v", err)
	}
	if metadata == nil || metadata.Camera == nil || *metadata.Camera != "FakeCam" {
		t.Fatalf("Parse() after restart = %+v, want camera FakeCam", metadata)
	}

	mime, err := MIMEType("photo.jpg")
	if err != nil {
		t.Fatalf("MIMEType() after restart error: %v", err)
	}
	if mime != "image/jpeg" {
		t.Fatalf("MIMEType() after restart = %q, want image/jpeg", mime)
	}
}

func TestMIMEType(t *testing.T) {
	resetForTest()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	filename := "./test_data/bird.jpg"

	mime, err := MIMEType(filename)
	if err != nil {
		t.Fatalf("MIMEType() returns an error: %v", err)
	}

	if mime == "" {
		t.Errorf("MIMEType(%q) should not return an empty string", filename)
	}
}

func fileModifyDateLiteralInUTC(t *testing.T, file string) time.Time {
	fstat, err := os.Stat(file)
	if err != nil {
		t.Fatalf("os.Stat(%q) error: %v", file, err)
	}

	ret := fstat.ModTime().Truncate(time.Second)
	_, offset := ret.Zone()
	ret = ret.Add(time.Duration(offset) * time.Second).UTC()

	return ret
}

func mustParseNoTimeZone(t *testing.T, timeStr string) time.Time {
	layout := "2006:01:02 15:04:05.999"
	ret, err := time.ParseInLocation(layout, timeStr, time.UTC)
	if err != nil {
		t.Fatalf("time.Parse(%q) error: %v", timeStr, err)
	}
	return ret
}

func TestSamplesTime(t *testing.T) {
	resetForTest()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	tests := []struct {
		file          string
		wantTime      time.Time
		wantOffsetSec *int
	}{
		{"./test_data/sample1.heif", fileModifyDateLiteralInUTC(t, "./test_data/sample1.heif"), nil},
		{"./test_data/sample1_nef.jpg", mustParseNoTimeZone(t, "2008:03:15 07:44:21.49"), new(-7 * 60 * 60)},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			metadata, err := Parse(tc.file)
			if err != nil {
				t.Fatalf("Parse(%q) returns an error: %v", tc.file, err)
			}

			if metadata == nil {
				t.Fatalf("metadata == nil")
			}
			if metadata.DateShot == nil {
				t.Fatalf("metadata.DateShot == nil")
			}

			if got := *metadata.DateShot; !got.Equal(tc.wantTime) {
				t.Errorf("metadata.DateShot = %v, want: %v", got, tc.wantTime)
			}

			if tc.wantOffsetSec == nil {
				if got := metadata.OffsetSecShot; got != nil {
					t.Errorf("metadata.OffsetSecShot = %+v, want: nil", got)
				}
			} else {
				if got, want := *metadata.OffsetSecShot, *tc.wantOffsetSec; got != want {
					t.Errorf("metadata.OffsetSecShot = &(%v), want: &(%v)", got, want)
				}
			}
		})
	}
}

const (
	fakeExiftoolControlEnv = "EXIFTOOL_FAKE_CONTROL"
	fakeExiftoolPIDEnv     = "EXIFTOOL_FAKE_PID"
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

func mustDeadline(t *testing.T, pidPath string, fn func() error) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v, want deadline exceeded", err)
		}
	case <-time.After(queryReturnBound):
		if pid := readPID(pidPath); pid > 0 {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
		select {
		case <-done:
		case <-time.After(queryReturnBound):
		}
		t.Fatal("query blocked past the deadline")
	}
}

func waitPID(t *testing.T, path string, except int) int {
	t.Helper()

	deadline := time.Now().Add(queryReturnBound)
	for {
		pid := readPID(path)
		if pid > 0 && pid != except {
			return pid
		}
		if time.Now().After(deadline) {
			t.Fatalf("exiftool pid file %s did not update from %d", path, except)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func readPID(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}

	return pid
}

func processGone(pid int) bool {
	return syscall.Kill(pid, 0) != nil
}
