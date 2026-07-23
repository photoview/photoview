package queue

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/buckket/go-blurhash"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/externaltools/exiftool"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	ignore "github.com/sabhiram/go-gitignore"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/gorm"
)

// worker owns everything needed to actually do the work: a database
// connection, and the tools it uses to process media. exif is exclusive to
// this worker (exiftool's process protocol does not support concurrent use);
// ffmpeg/magick/faces are stateless-per-call or already internally
// synchronized, so they're shared across all workers.
type worker struct {
	// ctx carries this worker's "worker_id" logging attribute (via
	// log.WithAttrs); every log call the worker makes uses it (or a further
	// derived context that also has the current task's media title on it),
	// instead of a bare nil context.
	ctx context.Context

	db     *gorm.DB
	exif   *exiftool.Exiftool
	ffmpeg *executable_worker.FfmpegCli
	magick *executable_worker.MagickWand
	faces  face_detection.FaceDetector
}

// activeWorkerCount tracks how many Workers are currently alive: incremented
// in newWorker, decremented in close(). A worker's worker_id is its count at
// the moment it started, so it reflects concurrency rather than a
// forever-growing lifetime counter.
var activeWorkerCount atomic.Int64

// newWorker always succeeds: if the worker's own exiftool process fails to
// start, that failure is reported (and gracefully degraded around) at the
// point of use in process(), the same way any other scanner error is - it
// does not prevent this worker from handling everything else.
func newWorker(parentCtx context.Context, db *gorm.DB) *worker {
	ctx := log.WithAttrs(context.WithoutCancel(parentCtx), "worker_id", activeWorkerCount.Add(1))

	exif, err := exiftool.New()
	if err != nil {
		scanner_utils.ScannerError(ctx, "start exiftool for scanner worker failed: %s", err)
		exif = nil
	}

	return &worker{
		ctx:    ctx,
		db:     db,
		exif:   exif,
		ffmpeg: executable_worker.Ffmpeg,
		magick: executable_worker.Magick,
		faces:  face_detection.GlobalFaceDetector,
	}
}

func (w *worker) close() {
	defer activeWorkerCount.Add(-1)

	if w.exif != nil {
		if err := w.exif.Close(); err != nil {
			log.Warn(w.ctx, "close scanner worker exiftool process failed", "error", err)
		}
	}
}

// processMedia runs a task through all four phases, reporting the outcome to
// its albumState. It never panics/returns an error to the caller - every
// failure is logged and notified (scanner_utils.ScannerError), matching how
// the rest of the scanner reports problems.
func (w *worker) processMedia(t *task) {
	ctx := w.ctx

	hasChanged := false
	defer func() {
		// t.info.media may already be set even if gather returned an error
		// (e.g. it failed on a later step, after determining/creating the
		// media row) - CompleteMedia still needs that ID so cleanup doesn't
		// mistake a real, existing media for stale.
		t.albumState.CompleteMedia(ctx, t.info.media, hasChanged)
		if t.done != nil {
			close(t.done)
		}
	}()

	info, err := w.gather(ctx, t)
	t.info = info
	if err != nil {
		scanner_utils.ScannerError(ctx, "gather info (%s): %s", t.path, err)
		return
	}

	if info.media != nil {
		ctx = log.WithAttrs(ctx, "media_title", info.media.Title)
	}

	if info.skip {
		return
	}

	t.plan = decide(info)
	if t.plan.isNewMedia {
		t.albumState.NotifyFoundNewMedia(t.path)
	}

	if t.plan.nothingToDo() {
		return
	}

	result, err := w.process(ctx, t)
	if err != nil {
		scanner_utils.ScannerError(ctx, "process media (%s): %s", t.path, err)
		return
	}
	t.result = result

	if err := w.persist(ctx, t); err != nil {
		scanner_utils.ScannerError(ctx, "save media to database (%s): %s", t.path, err)
		return
	}

	hasChanged = true
}

// gather (phase 1) reads the filesystem and database to establish every fact
// needed about this candidate file, exactly once.
func (w *worker) gather(ctx context.Context, t *task) (gatheredInfo, error) {
	var info gatheredInfo

	mediaType := t.cache.GetMediaType(t.path)
	info.mediaType = mediaType

	var albumIgnoreLines []string
	if lines := t.cache.GetAlbumIgnore(t.album.Path); lines != nil {
		albumIgnoreLines = *lines
	}
	albumIgnore := ignore.CompileIgnoreLines(albumIgnoreLines...)
	if albumIgnore.MatchesPath(path.Base(t.path)) {
		info.skip = true
		return info, nil
	}

	if !mediaType.IsSupported() {
		info.skip = true
		return info, nil
	}

	if utils.EnvDisableRawProcessing.GetBool() {
		if !mediaType.IsWebCompatible() {
			info.skip = true
			return info, nil
		}
	} else if mediaType.IsWebCompatible() {
		if _, existed := media_type.FindRawCounterpart(t.path); existed {
			info.skip = true
			return info, nil
		}
	}

	info.isVideo = mediaType.IsVideo()

	media, isNew, err := w.findOrCreateMedia(t.path, t.album.ID, mediaType)
	if err != nil {
		return info, err
	}
	info.media = media
	info.isNewMedia = isNew

	info.existingURLs = map[models.MediaPurpose]*models.MediaURL{}
	if !isNew {
		var urls []*models.MediaURL
		if err := w.db.Where("media_id = ?", media.ID).Find(&urls).Error; err != nil {
			return info, err
		}
		for i := range urls {
			info.existingURLs[urls[i].Purpose] = urls[i]
		}

		cachePath, err := media.CachePath()
		if err != nil {
			return info, err
		}
		if url, ok := info.existingURLs[models.PhotoThumbnail]; ok {
			info.thumbnailMissing = !scanner_utils.FileExists(path.Join(cachePath, url.MediaName))
		}
		if url, ok := info.existingURLs[models.PhotoHighRes]; ok {
			info.highresMissing = !scanner_utils.FileExists(path.Join(cachePath, url.MediaName))
		}
		if url, ok := info.existingURLs[models.VideoWeb]; ok {
			info.webVideoMissing = !scanner_utils.FileExists(path.Join(cachePath, url.MediaName))
		}
		if url, ok := info.existingURLs[models.VideoThumbnail]; ok {
			info.videoThumbMissing = !scanner_utils.FileExists(path.Join(cachePath, url.MediaName))
		}
	}

	if !info.isVideo && !mediaType.IsWebCompatible() {
		sidecarPath := t.path + ".xmp"
		if scanner_utils.FileExists(sidecarPath) {
			info.sidecarPath = &sidecarPath
			info.sidecarHash = hashFile(ctx, sidecarPath)
		}

		if counterpart, ok := media_type.FindWebCounterpart(t.path); ok {
			info.webCounterpart = counterpart
			info.hasCounterpart = true
		}
	}

	return info, nil
}

func (w *worker) findOrCreateMedia(mediaPath string, albumID int, mediaType media_type.MediaType) (*models.Media, bool, error) {
	var existing []*models.Media
	if err := w.db.Where("path_hash = ?", models.MD5Hash(mediaPath)).Find(&existing).Error; err != nil {
		return nil, false, err
	}
	if len(existing) > 0 {
		return existing[0], false, nil
	}

	stat, err := os.Stat(mediaPath)
	if err != nil {
		return nil, false, err
	}

	mediaTypeText := models.MediaTypePhoto
	if mediaType.IsVideo() {
		mediaTypeText = models.MediaTypeVideo
	}

	media := &models.Media{
		Title:    path.Base(mediaPath),
		Path:     mediaPath,
		AlbumID:  albumID,
		Type:     mediaTypeText,
		DateShot: stat.ModTime(),
	}

	if err := w.db.Create(media).Error; err != nil {
		return nil, false, err
	}

	return media, true, nil
}

func hashFile(ctx context.Context, p string) *string {
	f, err := os.Open(p)
	if err != nil {
		log.Warn(ctx, "hash file failed", "path", p, "error", err)
		return nil
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Warn(ctx, "hash file failed", "path", p, "error", err)
		return nil
	}

	hash := hex.EncodeToString(h.Sum(nil))
	return &hash
}

// process (phase 3) calls out to exiftool/ffprobe/ffmpeg/imagick/face
// detection/blurhash to compute results. It performs no database writes.
func (w *worker) process(ctx context.Context, t *task) (workResult, error) {
	info := t.info
	plan := t.plan
	var result workResult

	cachePath, err := info.media.CachePath()
	if err != nil {
		return result, fmt.Errorf("cache directory error: %w", err)
	}

	if plan.needExif {
		if w.exif == nil {
			scanner_utils.ScannerError(ctx, "exif unavailable, skip parsing exif for %s", t.path)
		} else if exifData, err := w.parseExif(t.path); err != nil {
			log.Warn(ctx, "parse exif failed", "path", t.path, "error", err)
		} else if exifData != nil {
			result.exif = exifData
			if exifData.DateShot != nil {
				result.dateShot = exifData.DateShot
			}
		}
	}

	encData := media_encoding.NewEncodeMediaData(info.media)
	if info.hasCounterpart {
		counterpart := info.webCounterpart
		encData.CounterpartPath = &counterpart
	}

	if !info.isVideo {
		if err := w.processPhoto(t, &encData, cachePath, &result); err != nil {
			return result, err
		}
	} else {
		if err := w.processVideo(ctx, t, cachePath, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (w *worker) processPhoto(t *task, encData *media_encoding.EncodeMediaData, cachePath string, result *workResult) error {
	info := t.info
	plan := t.plan
	baseImagePath := t.path

	if plan.needHighres {
		highresName := t.path
		if url, ok := info.existingURLs[models.PhotoHighRes]; ok {
			highresName = url.MediaName
		} else {
			highresName = generateUniqueMediaNamePrefixed("highres", t.path, ".jpg")
		}
		highresPath := path.Join(cachePath, highresName)

		if err := encData.EncodeHighRes(highresPath); err != nil {
			return fmt.Errorf("encode highres: %w", err)
		}

		dim, err := media_encoding.GetPhotoDimensions(highresPath)
		if err != nil {
			return err
		}
		stat, err := os.Stat(highresPath)
		if err != nil {
			return fmt.Errorf("stat highres: %w", err)
		}

		result.highres = &encodedFile{path: highresPath, name: highresName, width: dim.Width, height: dim.Height, fileSize: stat.Size(), contentType: "image/jpeg"}
		baseImagePath = highresPath
	} else if url, ok := info.existingURLs[models.PhotoHighRes]; ok {
		baseImagePath = path.Join(cachePath, url.MediaName)
	}

	if plan.needThumbnail {
		thumbnailName := ""
		if url, ok := info.existingURLs[models.PhotoThumbnail]; ok {
			thumbnailName = url.MediaName
		} else {
			thumbnailName = generateUniqueMediaNamePrefixed("thumbnail", t.path, ".jpg")
		}
		thumbPath := path.Join(cachePath, thumbnailName)

		dim, err := media_encoding.EncodeThumbnail(w.db, baseImagePath, thumbPath)
		if err != nil {
			return fmt.Errorf("encode thumbnail: %w", err)
		}
		stat, err := os.Stat(thumbPath)
		if err != nil {
			return fmt.Errorf("stat thumbnail: %w", err)
		}

		result.thumbnail = &encodedFile{path: thumbPath, name: thumbnailName, width: dim.Width, height: dim.Height, fileSize: stat.Size(), contentType: "image/jpeg"}
	}

	if plan.needOriginal {
		contentType, err := encData.ContentType()
		if err != nil {
			return err
		}
		dim, err := media_encoding.GetPhotoDimensions(t.path)
		if err != nil {
			return err
		}
		stat, err := os.Stat(t.path)
		if err != nil {
			return fmt.Errorf("stat original photo: %w", err)
		}

		result.original = &encodedFile{path: t.path, name: generateUniqueMediaName(t.path), width: dim.Width, height: dim.Height, fileSize: stat.Size(), contentType: contentType.String()}
	}

	if plan.needBlurhash {
		thumbPath := ""
		if result.thumbnail != nil {
			thumbPath = result.thumbnail.path
		} else if url, ok := info.existingURLs[models.PhotoThumbnail]; ok {
			thumbPath = path.Join(cachePath, url.MediaName)
		}
		if thumbPath != "" {
			hash, err := generateBlurhash(thumbPath)
			if err != nil {
				return fmt.Errorf("generate blurhash: %w", err)
			}
			result.blurhash = &hash
		}
	}

	if plan.sidecarChanged {
		result.sidecarPath = info.sidecarPath
		result.sidecarHash = info.sidecarHash
	}

	return nil
}

func (w *worker) processVideo(ctx context.Context, t *task, cachePath string, result *workResult) error {
	info := t.info
	plan := t.plan

	probeData, err := probeVideo(t.path)
	if err != nil {
		return err
	}

	if plan.needOriginal {
		stream := probeData.FirstVideoStream()
		if stream == nil {
			return fmt.Errorf("could not get video stream from metadata (%s)", t.path)
		}
		stat, err := os.Stat(t.path)
		if err != nil {
			return fmt.Errorf("stat original video: %w", err)
		}
		result.original = &encodedFile{path: t.path, name: generateUniqueMediaName(t.path), width: stream.Width, height: stream.Height, fileSize: stat.Size(), contentType: info.mediaType.String()}
	}

	if plan.needWebVideo {
		webVideoName := generateUniqueMediaNamePrefixed("web_video", t.path, ".mp4")
		webVideoPath := path.Join(cachePath, webVideoName)

		if err := w.ffmpeg.EncodeMp4(t.path, webVideoPath); err != nil {
			return fmt.Errorf("encode web video: %w", err)
		}

		webProbe, err := probeVideo(webVideoPath)
		if err != nil {
			return err
		}
		stream := webProbe.FirstVideoStream()
		if stream == nil {
			return fmt.Errorf("could not get video stream from metadata (%s)", webVideoPath)
		}
		stat, err := os.Stat(webVideoPath)
		if err != nil {
			return fmt.Errorf("stat web video: %w", err)
		}

		result.webVideo = &encodedFile{path: webVideoPath, name: webVideoName, width: stream.Width, height: stream.Height, fileSize: stat.Size(), contentType: "video/mp4"}
	}

	if plan.needVideoThumb {
		videoThumbName := ""
		if url, ok := info.existingURLs[models.VideoThumbnail]; ok {
			videoThumbName = url.MediaName
		} else {
			videoThumbName = generateUniqueMediaNamePrefixed("video_thumb", t.path, ".jpg")
		}
		thumbPath := path.Join(cachePath, videoThumbName)

		if err := w.ffmpeg.EncodeVideoThumbnail(t.path, thumbPath, probeData); err != nil {
			return fmt.Errorf("encode video thumbnail: %w", err)
		}

		dim, err := media_encoding.GetPhotoDimensions(thumbPath)
		if err != nil {
			return err
		}
		stat, err := os.Stat(thumbPath)
		if err != nil {
			return fmt.Errorf("stat video thumbnail: %w", err)
		}

		result.videoThumb = &encodedFile{path: thumbPath, name: videoThumbName, width: dim.Width, height: dim.Height, fileSize: stat.Size(), contentType: "image/jpeg"}
	}

	if plan.needBlurhash {
		thumbPath := ""
		if result.videoThumb != nil {
			thumbPath = result.videoThumb.path
		} else if url, ok := info.existingURLs[models.VideoThumbnail]; ok {
			thumbPath = path.Join(cachePath, url.MediaName)
		}
		if thumbPath != "" {
			hash, err := generateBlurhash(thumbPath)
			if err != nil {
				return fmt.Errorf("generate blurhash: %w", err)
			}
			result.blurhash = &hash
		}
	}

	if plan.needVideoMeta {
		vm, err := buildVideoMetadata(probeData)
		if err != nil {
			log.Warn(ctx, "read video metadata failed", "path", t.path, "error", err)
		} else {
			result.videoMeta = vm
		}
	}

	return nil
}

func (w *worker) parseExif(path string) (*models.MediaEXIF, error) {
	var values struct {
		exiftool.PhotoMeta
		exiftool.TimeAll
		exiftool.GPS
	}
	if err := w.exif.QueryJSONTagsByNumber(path, &values); err != nil {
		return nil, err
	}
	values.PhotoMeta.SanitizeFloats()

	ret := &models.MediaEXIF{
		Camera:          values.Model,
		Maker:           values.Make,
		Lens:            values.LensModel,
		Iso:             values.ISO,
		Flash:           values.Flash,
		Orientation:     values.Orientation,
		ExposureProgram: values.ExposureProgram,
		Exposure:        values.ExposureTime,
		Aperture:        values.Aperture,
		FocalLength:     values.FocalLength,
		Description:     values.ImageDescription,
	}

	dateShot := values.TimeAll.TimeInLocal()
	if !dateShot.IsZero() {
		d := dateShot
		ret.DateShot = &d
	}

	if offsetSec, ok := values.TimeAll.OffsetSecs(dateShot); ok {
		ret.OffsetSecShot = &offsetSec
	}

	if values.GPS.IsValid() {
		ret.GPSLatitude = values.GPS.GPSLatitude
		ret.GPSLongitude = values.GPS.GPSLongitude
	}

	return ret, nil
}

func probeVideo(videoPath string) (*ffprobe.ProbeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), utils.MediaProbeTimeout())
	defer cancel()

	data, err := ffprobe.ProbeURL(ctx, videoPath)
	if err != nil {
		return nil, fmt.Errorf("could not read video metadata (%s): %w", path.Base(videoPath), err)
	}

	return data, nil
}

func buildVideoMetadata(data *ffprobe.ProbeData) (*models.VideoMetadata, error) {
	stream := data.FirstVideoStream()
	if stream == nil {
		return nil, fmt.Errorf("could not get video stream from metadata")
	}

	audio := data.FirstAudioStream()
	var audioText string
	switch {
	case audio == nil || audio.Channels == 0:
		audioText = "No audio"
	case audio.Channels == 1:
		audioText = "Mono audio"
	case audio.Channels == 2:
		audioText = "Stereo audio"
	default:
		audioText = fmt.Sprintf("Audio (%d channels)", audio.Channels)
	}

	return &models.VideoMetadata{
		Width:        stream.Width,
		Height:       stream.Height,
		Duration:     data.Format.DurationSeconds,
		Codec:        &stream.CodecLongName,
		Framerate:    parseFramerate(stream.AvgFrameRate),
		Bitrate:      &stream.BitRate,
		ColorProfile: &stream.Profile,
		Audio:        &audioText,
	}, nil
}

func parseFramerate(s string) *float64 {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil
	}
	num, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil
	}
	den, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || den == 0 {
		return nil
	}
	r := float64(num) / float64(den)
	return &r
}

func generateBlurhash(imagePath string) (string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	imageData, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	return blurhash.Encode(4, 3, imageData)
}

func generateUniqueMediaName(mediaPath string) string {
	filename := path.Base(mediaPath)
	ext := path.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	name := fmt.Sprintf("%s_%s", baseName, utils.GenerateToken())
	return models.SanitizeMediaName(name) + ext
}

func generateUniqueMediaNamePrefixed(prefix string, mediaPath string, extension string) string {
	name := fmt.Sprintf("%s_%s_%s", prefix, path.Base(mediaPath), utils.GenerateToken())
	return models.SanitizeMediaName(name) + extension
}

// persist (phase 4) writes everything process() computed to the database in
// a single transaction.
func (w *worker) persist(ctx context.Context, t *task) error {
	info := t.info
	plan := t.plan
	result := t.result

	return w.db.Transaction(func(tx *gorm.DB) error {
		media := info.media

		if result.dateShot != nil {
			media.DateShot = *result.dateShot
		}
		if result.blurhash != nil {
			media.Blurhash = result.blurhash
		}
		if plan.sidecarChanged {
			media.SideCarPath = result.sidecarPath
			media.SideCarHash = result.sidecarHash
		}

		if err := tx.Save(media).Error; err != nil {
			return fmt.Errorf("save media: %w", err)
		}

		if result.exif != nil {
			if err := tx.Model(media).Association("Exif").Replace(result.exif); err != nil {
				return fmt.Errorf("save media exif: %w", err)
			}
		}

		if result.videoMeta != nil {
			media.VideoMetadata = result.videoMeta
			if err := tx.Save(media).Error; err != nil {
				return fmt.Errorf("save video metadata: %w", err)
			}
		}

		for purpose, file := range map[models.MediaPurpose]*encodedFile{
			models.MediaOriginal:  result.original,
			models.PhotoThumbnail: result.thumbnail,
			models.PhotoHighRes:   result.highres,
			models.VideoWeb:       result.webVideo,
			models.VideoThumbnail: result.videoThumb,
		} {
			if file == nil {
				continue
			}
			if err := upsertMediaURL(tx, media, purpose, file, info.existingURLs[purpose]); err != nil {
				return fmt.Errorf("save media url (%s): %w", purpose, err)
			}
		}

		if plan.needFaces && w.faces != nil {
			if err := w.faces.DetectFaces(tx, media); err != nil {
				scanner_utils.ScannerError(ctx, "detect faces (%s): %s", t.path, err)
			}
		}

		return nil
	})
}

func upsertMediaURL(tx *gorm.DB, media *models.Media, purpose models.MediaPurpose, file *encodedFile, existing *models.MediaURL) error {
	if existing != nil {
		existing.Width = file.width
		existing.Height = file.height
		existing.FileSize = file.fileSize
		return tx.Save(existing).Error
	}

	url := &models.MediaURL{
		MediaID:     media.ID,
		MediaName:   file.name,
		Width:       file.width,
		Height:      file.height,
		Purpose:     purpose,
		ContentType: file.contentType,
		FileSize:    file.fileSize,
	}
	return tx.Create(url).Error
}
