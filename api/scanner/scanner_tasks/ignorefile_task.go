package scanner_tasks

import (
	"io/fs"
	"log"

	"github.com/photoview/photoview/api/scanner/scanner_task"
	ignore "github.com/sabhiram/go-gitignore"
)

type IgnorefileTask struct {
	scanner_task.ScannerTaskBase
}

type ignorefileTaskKey string

const albumIgnoreKey ignorefileTaskKey = "album_ignore_key"

func getAlbumIgnore(ctx scanner_task.TaskContext) *ignore.GitIgnore {
	return ctx.Value(albumIgnoreKey).(*ignore.GitIgnore)
}

func (t IgnorefileTask) BeforeScanAlbum(ctx scanner_task.TaskContext) (scanner_task.TaskContext, error) {
	albumIgnore := ignore.CompileIgnoreLines(*ctx.GetCache().GetAlbumIgnore(ctx.GetAlbum().Path)...)
	return ctx.WithValue(albumIgnoreKey, albumIgnore), nil
}

func (t IgnorefileTask) MediaFound(ctx scanner_task.TaskContext, fileInfo fs.FileInfo, mediaPath string) (bool, error) {

	// Match file against ignore data
	if getAlbumIgnore(ctx).MatchesPath(fileInfo.Name()) {
		log.Printf("File %s ignored\n", fileInfo.Name())
		return true, nil
	}

	return false, nil
}
