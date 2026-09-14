package utils_test

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/utils"
	"github.com/stretchr/testify/assert"
)

func TestPeriodicScanInterval(t *testing.T) {
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(utils.EnvPeriodicScanInterval.GetName(), tc.value)
			}

			interval, ok := utils.PeriodicScanInterval()

			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, interval)
		})
	}
}
