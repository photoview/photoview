package utils_test

import (
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/photoview/photoview/api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxSeconds = int64(math.MaxInt64) / int64(time.Second)

func TestPeriodicScanInterval(t *testing.T) {
	name := utils.EnvPeriodicScanInterval.GetName()

	tests := []struct {
		name     string
		value    string
		setEnv   bool
		expected time.Duration
		ok       bool
	}{
		{name: "unset falls back to the database"},
		{name: "empty falls back to the database", value: "", setEnv: true},
		{name: "interval in seconds", value: "3600", setEnv: true, expected: time.Hour, ok: true},
		{name: "zero disables periodic scanning", value: "0", setEnv: true, ok: true},
		{name: "not a number falls back", value: "hourly", setEnv: true},
		{name: "negative falls back", value: "-1", setEnv: true},
		{
			name:     "largest representable interval",
			value:    strconv.FormatInt(maxSeconds, 10),
			setEnv:   true,
			expected: time.Duration(maxSeconds) * time.Second,
			ok:       true,
		},
		{name: "overflowing interval falls back", value: strconv.FormatInt(maxSeconds+1, 10), setEnv: true},
		{name: "beyond int64 falls back", value: "99999999999999999999", setEnv: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Setenv restores whatever the process inherited; unsetting afterwards
			// keeps the "unset" cases from reading the host configuration.
			t.Setenv(name, tc.value)
			if !tc.setEnv {
				require.NoError(t, os.Unsetenv(name))
			}

			interval, ok := utils.PeriodicScanInterval()

			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, interval)
		})
	}
}
