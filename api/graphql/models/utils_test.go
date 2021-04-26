package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
)

func TestMD5Hash(t *testing.T) {
	assert.Equal(t, "5eb63bbbe01eeed093cb22bb8f5acdc3", models.MD5Hash("hello world"))
}
