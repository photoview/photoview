package queue

import (
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
)

// task holds all state for processing a single candidate media file as it
// moves through gather -> decide -> process -> persist. task itself carries
// no behavior; all four phases are implemented as worker methods.
type task struct {
	album      *models.Album
	cache      *scanner_cache.AlbumScannerCache
	albumState *albumState
	path       string

	// done is only set for a task submitted directly via ProcessMedia, which
	// needs to block until this specific task finishes. Tasks produced by
	// expandAlbum leave it nil - nothing waits on those individually.
	done chan struct{}

	info   gatheredInfo
	plan   workPlan
	result workResult
}

// newTask builds a task for one candidate media file. done should be nil for
// tasks produced by expandAlbum (nothing waits on those individually), or a
// fresh channel for a task submitted directly via ProcessMedia, which blocks
// on it until this specific task finishes.
func newTask(album *models.Album, cache *scanner_cache.AlbumScannerCache, albumState *albumState, path string, done chan struct{}) *task {
	return &task{
		album:      album,
		cache:      cache,
		albumState: albumState,
		path:       path,
		done:       done,
	}
}

// gatheredInfo is the output of worker.gather (phase 1).
type gatheredInfo struct {
	skip bool

	mediaType media_type.MediaType
	isVideo   bool

	media      *models.Media
	isNewMedia bool

	existingURLs map[models.MediaPurpose]*models.MediaURL

	thumbnailMissing  bool
	highresMissing    bool
	webVideoMissing   bool
	videoThumbMissing bool

	webCounterpart string
	hasCounterpart bool

	sidecarPath *string
	sidecarHash *string
}

// workPlan is the output of decide (phase 2): what work, if any, is needed.
type workPlan struct {
	isNewMedia bool

	needExif       bool
	needVideoMeta  bool
	needThumbnail  bool
	needHighres    bool
	needOriginal   bool
	needWebVideo   bool
	needVideoThumb bool
	sidecarChanged bool

	mediaChanged bool // a thumbnail/highres/web-video/video-thumbnail was (re)generated
	needFaces    bool
	needBlurhash bool
}

func (p workPlan) nothingToDo() bool {
	return !p.needExif && !p.needVideoMeta && !p.mediaChanged &&
		!p.needFaces && !p.needBlurhash && !p.needOriginal
}

// encodedFile describes a cache file that was generated/verified during process (phase 3).
type encodedFile struct {
	path        string
	name        string
	width       int
	height      int
	fileSize    int64
	contentType string
}

// workResult is the output of process (phase 3), persisted by worker.persist (phase 4).
type workResult struct {
	dateShot  *time.Time
	exif      *models.MediaEXIF
	videoMeta *models.VideoMetadata

	original   *encodedFile
	thumbnail  *encodedFile
	highres    *encodedFile
	webVideo   *encodedFile
	videoThumb *encodedFile

	blurhash *string

	sidecarPath *string
	sidecarHash *string
}

// decide is pure logic: given what gather found, decide what work is needed.
// It touches no tools, no filesystem, no database.
func decide(info gatheredInfo) workPlan {
	var p workPlan
	p.isNewMedia = info.isNewMedia

	p.needExif = info.isNewMedia
	p.needVideoMeta = info.isNewMedia && info.isVideo
	p.needOriginal = info.existingURLs[models.MediaOriginal] == nil &&
		(!info.isVideo || info.mediaType.IsWebCompatible())

	if !info.isVideo {
		p.needHighres = !info.mediaType.IsWebCompatible() &&
			(info.existingURLs[models.PhotoHighRes] == nil || info.highresMissing)
		p.needThumbnail = info.existingURLs[models.PhotoThumbnail] == nil || info.thumbnailMissing

		if !info.mediaType.IsWebCompatible() {
			p.sidecarChanged = sidecarChanged(info)
			if p.sidecarChanged && !info.isNewMedia {
				p.needHighres = true
				p.needThumbnail = true
			}
		}
	} else {
		p.needWebVideo = !info.mediaType.IsWebCompatible() &&
			(info.existingURLs[models.VideoWeb] == nil || info.webVideoMissing)
		p.needVideoThumb = info.existingURLs[models.VideoThumbnail] == nil || info.videoThumbMissing
	}

	p.mediaChanged = p.needThumbnail || p.needHighres || p.needWebVideo || p.needVideoThumb || p.sidecarChanged

	p.needFaces = p.mediaChanged && !info.isVideo
	p.needBlurhash = p.mediaChanged || (info.media != nil && info.media.Blurhash == nil)

	return p
}

// sidecarChanged assumes info.media is already set - decide() is only ever
// called with a gatheredInfo returned by a successful, non-skipped gather(),
// which always determines/creates the media row before reaching this check.
func sidecarChanged(info gatheredInfo) bool {
	if info.sidecarPath != nil {
		return info.media.SideCarHash == nil || *info.media.SideCarHash != *info.sidecarHash
	}

	// sidecar has been deleted since last scan
	return info.media.SideCarPath != nil
}
