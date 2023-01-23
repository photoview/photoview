package watcher_scanner

import (
	"github.com/fsnotify/fsnotify"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"sync"
)

var mainWatcherScanner *watcherScanner = nil

type watcherScanner struct {
	watcherChanged chan bool
	watcher        *fsnotify.Watcher
	mutex          *sync.Mutex
	db             *gorm.DB
}

func InitializeWatcherScanner(db *gorm.DB) error {
	if mainWatcherScanner != nil {
		panic("watcher scanner has already been initialized")
	}

	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	mainWatcherScanner = &watcherScanner{
		db:             db,
		watcherChanged: make(chan bool),
		watcher:        watcher,
		mutex:          &sync.Mutex{},
	}

	success, errs := mainWatcherScanner.addPathsToWatch()
	if len(errs) != 0 {
		log.Println("watcher errors")
		log.Println(errs)
		if !success {
			return errors.New("errors found during watcher scanner setup")
		}
	}

	go mainWatcherScanner.processWatchEvents()

	return nil
}

func (ws watcherScanner) addPathsToWatch() (bool, []error) {
	var allAlbumPaths []*models.Album

	if err := ws.db.Select("path").Find(&allAlbumPaths).Error; err != nil {
		return false, []error{errors.Wrap(err, "watcher scanner find albums query")}
	}

	errs := make([]error, 0)
	for _, album := range allAlbumPaths {
		// log.Println("add path", album.Path)
		err := ws.watcher.Add(album.Path)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "add path watcher scanner"))
		}
	}
	return len(allAlbumPaths) != len(errs), errs
}

func (ws watcherScanner) processWatchEvents() {
	log.Println("watching for events")
	for {
		select {
		case event, ok := <-ws.watcher.Events:
			if !ok {
				//log.Println("not ok", event)
				continue
			}
			log.Println("event:", event)
			var media *models.Media
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				scanner_queue.AddMediaToQueue(event.Name)
				log.Println("create event ", event.Name, event.Op)
			} else if event.Has(fsnotify.Remove) {
				ws.db.Preload(clause.Associations).Where("path = ?", event.Name).Find(&media)
				// todo: why doesnt exif cascade delete invoked ?
				if media.Exif != nil {
					ws.db.Delete(&media.Exif)
				}
				ws.db.Select(clause.Associations).Delete(&media)
				log.Println("remove event ", event.Name, event.Op)
			} else if event.Has(fsnotify.Rename) {
				log.Println("rename event ", event.Name, event.Op)
			}
		case err, ok := <-ws.watcher.Errors:
			if !ok {
				//log.Println("not ok, error", err)
				continue
			}
			log.Println("error:", err)
		}
	}
}
