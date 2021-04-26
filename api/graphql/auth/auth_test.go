package auth_test

import (
	"os"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.UnitTestRun(m))
}

func TestTokenFromBearer(t *testing.T) {

	testsValues := []struct {
		name   string
		bearer string
		out    string
		valid  bool
	}{
		{"Valid bearer", "Bearer ZY9YfxFa3TapSAD37XUBFryo", "ZY9YfxFa3TapSAD37XUBFryo", true},
		{"Case insensitive bearer", "bEaReR ZY9YfxFa3TapSAD37XUBFryo", "ZY9YfxFa3TapSAD37XUBFryo", true},
		{"Missing bearer start", "ZY9YfxFa3TapSAD37XUBFryo", "", false},
		{"Empty input", "", "", false},
		{"Invalid token value", "Bearer THIS_IS_INVALID", "", false},
	}

	for _, test := range testsValues {
		t.Run(test.name, func(t *testing.T) {
			token, err := auth.TokenFromBearer(&test.bearer)
			if test.valid {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, test.out, *token)
			} else {
				assert.Error(t, err)
				assert.Nil(t, token)
			}
		})
	}
}
