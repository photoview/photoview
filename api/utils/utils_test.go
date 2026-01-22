package utils_test

import (
	"path"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/spf13/afero"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

func TestIsDirSymlink(t *testing.T) {
	fs, _ := test_utils.FilesystemTest(t)

	symlinker, ok := fs.(afero.Symlinker)
	if !ok {
		t.Fatalf("filesystem does not support symlinks")
	}

	// Prepare a temporary directory for testing purposes
	dir, err := afero.TempDir(fs, "", "testing")
	if err != nil {
		t.Fatalf("unable to create temp directory for testing")
	}
	defer fs.RemoveAll(dir)

	// Create regular file
	_, err = fs.Create(path.Join(dir, "regular_file"))
	if err != nil {
		t.Fatalf("unable to create regular file for testing")
	}

	// Create directory
	err = fs.Mkdir(path.Join(dir, "directory"), 0755)
	if err != nil {
		t.Fatalf("unable to create directory for testing")
	}

	// Create symlink to regular file
	err = symlinker.SymlinkIfPossible(path.Join(dir, "regular_file"), path.Join(dir, "file_link"))
	if err != nil {
		t.Fatalf("unable to create file link for testing")
	}

	// Create symlink to directory
	err = symlinker.SymlinkIfPossible(path.Join(dir, "directory"), path.Join(dir, "dir_link"))
	if err != nil {
		t.Fatalf("unable to create dir link for testing")
	}

	// Execute the actual tests

	isDirLink, _ := utils.IsDirSymlink(fs, path.Join(dir, "regular_file"))
	if isDirLink {
		t.Error("Failed detection of regular file")
	}

	isDirLink, _ = utils.IsDirSymlink(fs, path.Join(dir, "directory"))
	if isDirLink {
		t.Error("Failed detection of directory")
	}

	isDirLink, _ = utils.IsDirSymlink(fs, path.Join(dir, "file_link"))
	if isDirLink {
		t.Error("Failed detection of link to regular file")
	}

	isDirLink, _ = utils.IsDirSymlink(fs, path.Join(dir, "dir_link"))
	if !isDirLink {
		t.Error("Failed detection of link to directory")
	}

	isDirLink, err = utils.IsDirSymlink(fs, path.Join(dir, "non_existant"))
	if err == nil {
		t.Error("Missing error for non-existant file")
	}
}
