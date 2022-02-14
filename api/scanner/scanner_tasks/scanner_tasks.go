package scanner_tasks

import (
	"io/fs"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
)

var allTasks []scanner_task.ScannerTask = []scanner_task.ScannerTask{
	NotificationTask{},
	IgnorefileTask{},
}

type scannerTasks struct {
	scanner_task.ScannerTaskBase
}

var Tasks scannerTasks = scannerTasks{}

type scannerTasksKey string

const (
	tasksSubContextsGlobal     scannerTasksKey = "tasks_sub_contexts_global"
	tasksSubContextsProcessing scannerTasksKey = "tasks_sub_contexts_processing"
)

func getSubContextsGlobal(ctx scanner_task.TaskContext) []scanner_task.TaskContext {
	return ctx.Value(tasksSubContextsGlobal).([]scanner_task.TaskContext)
}

func getSubContextsProcessing(ctx scanner_task.TaskContext) []scanner_task.TaskContext {
	return ctx.Value(tasksSubContextsGlobal).([]scanner_task.TaskContext)
}

func (t scannerTasks) BeforeScanAlbum(ctx scanner_task.TaskContext) (scanner_task.TaskContext, error) {
	subContexts := make([]scanner_task.TaskContext, len(allTasks))

	for i, task := range allTasks {
		var err error
		subContexts[i], err = task.BeforeScanAlbum(ctx)
		if err != nil {
			return ctx, err
		}
	}

	return ctx.WithValue(tasksSubContextsGlobal, subContexts), nil
}

func (t scannerTasks) MediaFound(ctx scanner_task.TaskContext, fileInfo fs.FileInfo, mediaPath string) (bool, error) {

	subContexts := getSubContextsGlobal(ctx)

	for i, task := range allTasks {
		skip, err := task.MediaFound(subContexts[i], fileInfo, mediaPath)

		if err != nil {
			return false, err
		}

		if skip {
			return true, nil
		}
	}

	return false, nil
}

func (t scannerTasks) AfterScanAlbum(ctx scanner_task.TaskContext, albumHadChanges bool) error {
	subContexts := getSubContextsGlobal(ctx)
	for i, task := range allTasks {
		err := task.AfterScanAlbum(subContexts[i], albumHadChanges)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t scannerTasks) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {
	subContexts := getSubContextsGlobal(ctx)
	for i, task := range allTasks {
		err := task.AfterMediaFound(subContexts[i], media, newMedia)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t scannerTasks) BeforeProcessMedia(ctx scanner_task.TaskContext, media *models.Media) (scanner_task.TaskContext, error) {
	subContexts := make([]scanner_task.TaskContext, len(allTasks))

	for i, task := range allTasks {
		var err error
		subContexts[i], err = task.BeforeProcessMedia(ctx, media)
		if err != nil {
			return ctx, err
		}
	}

	return ctx.WithValue(tasksSubContextsProcessing, subContexts), nil
}

func (t scannerTasks) ProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, mediaCachePath string) (bool, error) {
	subContexts := getSubContextsProcessing(ctx)
	didProcess := false
	for i, task := range allTasks {
		singleDidProcess, err := task.ProcessMedia(subContexts[i], mediaData, mediaCachePath)
		if err != nil {
			return false, err
		}

		if singleDidProcess {
			didProcess = true
		}
	}
	return didProcess, nil
}

func (t scannerTasks) AfterProcessMedia(ctx scanner_task.TaskContext, media *models.Media, didProcess bool, mediaIndex int, mediaTotal int) error {
	subContexts := getSubContextsProcessing(ctx)
	for i, task := range allTasks {
		err := task.AfterProcessMedia(subContexts[i], media, didProcess, mediaIndex, mediaTotal)
		if err != nil {
			return err
		}
	}
	return nil
}
