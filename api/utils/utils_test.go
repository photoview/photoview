package utils_test

import (
	"os"
	"path"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

func TestSanitizeShareLabel(t *testing.T) {
	tests := []struct {
		name     string
		label    *string
		expected *string
	}{
		{
			name: "nil label",
		},
		{
			name:     "control and format characters",
			label:    new("Fam\x00ily\r\n\t\u007f\u0085\u200b album"),
			expected: new("Family album"),
		},
		{
			name:     "surrounding whitespace",
			label:    new("\u00a0  Family album \t"),
			expected: new("Family album"),
		},
		{
			name:  "empty label",
			label: new(""),
		},
		{
			name:  "invisible characters only",
			label: new("\x00\r\n\t\u007f\u0085\u200b"),
		},
		{
			name:     "ordinary unicode",
			label:    new("  Семья 日本語  "),
			expected: new("Семья 日本語"),
		},
		{
			name:     "semantic format characters",
			label:    new("Family 👩‍💻 می‌روم"),
			expected: new("Family 👩‍💻 می‌روم"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, utils.SanitizeShareLabel(test.label))
		})
	}
}

func TestIsDirSymlink(t *testing.T) {
	test_utils.FilesystemTest(t)

	// Prepare a temporary directory for testing purposes
	dir, err := os.MkdirTemp("", "testing")
	if err != nil {
		t.Fatalf("unable to create temp directory for testing")
	}
	defer os.RemoveAll(dir)

	// Create regular file
	_, err = os.Create(path.Join(dir, "regular_file"))
	if err != nil {
		t.Fatalf("unable to create regular file for testing")
	}

	// Create directory
	err = os.Mkdir(path.Join(dir, "directory"), 0755)
	if err != nil {
		t.Fatalf("unable to create directory for testing")
	}

	// Create symlink to regular file
	err = os.Symlink(path.Join(dir, "regular_file"), path.Join(dir, "file_link"))
	if err != nil {
		t.Fatalf("unable to create file link for testing")
	}

	// Create symlink to directory
	err = os.Symlink(path.Join(dir, "directory"), path.Join(dir, "dir_link"))
	if err != nil {
		t.Fatalf("unable to create dir link for testing")
	}

	// Execute the actual tests

	isDirLink, _ := utils.IsDirSymlink(path.Join(dir, "regular_file"))
	if isDirLink {
		t.Error("Failed detection of regular file")
	}

	isDirLink, _ = utils.IsDirSymlink(path.Join(dir, "directory"))
	if isDirLink {
		t.Error("Failed detection of directory")
	}

	isDirLink, _ = utils.IsDirSymlink(path.Join(dir, "file_link"))
	if isDirLink {
		t.Error("Failed detection of link to regular file")
	}

	isDirLink, _ = utils.IsDirSymlink(path.Join(dir, "dir_link"))
	if !isDirLink {
		t.Error("Failed detection of link to directory")
	}

	isDirLink, err = utils.IsDirSymlink(path.Join(dir, "non_existant"))
	if err == nil {
		t.Error("Missing error for non-existant file")
	}
}
