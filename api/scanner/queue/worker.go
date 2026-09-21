package queue

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
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
		ctx = log.WithAttrs(ctx, "media_path", info.media.Path)
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
		cleanupPendingCache(ctx, result.pendingCachePath)
		return
	}
	t.result = result

	if err := w.persist(ctx, t); err != nil {
		scanner_utils.ScannerError(ctx, "save media to database (%s): %s", t.path, err)
		cleanupPendingCache(ctx, result.pendingCachePath)
		return
	}

	hasChanged = true
}

// cleanupPendingCache removes a new media's scratch cache directory after
// process() or persist() failed partway - it's only ever set (non-empty)
// for a brand-new media, and only ever renamed away on a fully successful
// persist(), so a failure anywhere before that leaves it right where
// process() created it.
func cleanupPendingCache(ctx context.Context, pendingCachePath string) {
	if pendingCachePath == "" {
		return
	}
	if err := os.RemoveAll(pendingCachePath); err != nil {
		log.Warn(ctx, "clean up pending cache directory failed", "path", pendingCachePath, "error", err)
	}
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

	media, isNew, err := w.findOrBuildMedia(t.path, t.album.ID, mediaType)
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
			hash, err := hashFile(sidecarPath)
			if err != nil {
				return info, fmt.Errorf("hash sidecar file: %w", err)
			}
			info.sidecarHash = hash
		}

		if counterpart, ok := media_type.FindWebCounterpart(t.path); ok {
			info.webCounterpart = counterpart
			info.hasCounterpart = true
		}
	}

	return info, nil
}

// findOrBuildMedia looks up an existing Media row by path_hash (read-only -
// it never writes). If none exists, it builds an in-memory, unpersisted
// Media (ID left at zero) with every field gather() can already determine;
// the actual INSERT happens later, in persist(), once processing has
// produced something worth making visible alongside it.
func (w *worker) findOrBuildMedia(mediaPath string, albumID int, mediaType media_type.MediaType) (*models.Media, bool, error) {
	var existing models.Media
	err := w.db.Where("path_hash = ?", models.MD5Hash(mediaPath)).First(&existing).Error
	switch {
	case err == nil:
		return &existing, false, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, false, err
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

	return media, true, nil
}

func hashFile(p string) (*string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	hash := hex.EncodeToString(h.Sum(nil))
	return &hash, nil
}

// process (phase 3) calls out to exiftool/ffprobe/ffmpeg/imagick/face
// detection/blurhash to compute results. It performs no database writes.
func (w *worker) process(ctx context.Context, t *task) (workResult, error) {
	info := t.info
	plan := t.plan
	var result workResult

	var cachePath string
	var err error
	if info.isNewMedia {
		// The token suffix matters: two workers can concurrently decide the
		// same brand-new path needs processing (e.g. the same physical album
		// discovered twice because two users share it) - gather() is
		// read-only now, so nothing claims the path until persist() upserts
		// it. Without a random suffix here, both workers would write into
		// (and one would rename away from under the other) the same
		// directory.
		cachePath, err = utils.PendingCachePathForMedia(info.media.AlbumID, "tmp-"+models.MD5Hash(t.path)+"-"+utils.GenerateToken())
		if err != nil {
			return result, fmt.Errorf("pending cache directory error: %w", err)
		}
		result.pendingCachePath = cachePath
	} else {
		cachePath, err = info.media.CachePath()
		if err != nil {
			return result, fmt.Errorf("cache directory error: %w", err)
		}
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
		highresName := generateUniqueMediaNamePrefixed("highres", t.path, ".jpg")
		if url, ok := info.existingURLs[models.PhotoHighRes]; ok {
			highresName = url.MediaName
		}
		highresPath := path.Join(cachePath, highresName)

		if err := encodeAtomically(highresPath, encData.EncodeHighRes); err != nil {
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

		var dim media_encoding.Dimension
		err := encodeAtomically(thumbPath, func(tmpPath string) error {
			var err error
			dim, err = media_encoding.EncodeThumbnail(w.db, baseImagePath, tmpPath)
			return err
		})
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

		if err := encodeAtomically(webVideoPath, func(tmpPath string) error {
			return w.ffmpeg.EncodeMp4(t.path, tmpPath)
		}); err != nil {
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

		if err := encodeAtomically(thumbPath, func(tmpPath string) error {
			return w.ffmpeg.EncodeVideoThumbnail(t.path, tmpPath, probeData)
		}); err != nil {
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

// encodeAtomically calls encode with a fresh temporary path next to outPath,
// then renames it into place - so a partial write (OOM kill, full disk,
// crash) never leaves outPath itself corrupted or truncated. This matters
// even when outPath doesn't exist yet: without it, a later scan's
// scanner_utils.FileExists check would see the half-written file as already
// present and never retry it. The token is prefixed onto outPath's full
// filename (extension included, e.g. "tmp-<rand>.video.mp4") rather than
// appended after it, since ffmpeg infers its output container from the
// destination filename's extension - EncodeMp4/EncodeVideoThumbnail are
// never told the format explicitly. encode must write its result to the
// tmpPath it's given, not to outPath directly.
func encodeAtomically(outPath string, encode func(tmpPath string) error) error {
	tmpPath := path.Join(path.Dir(outPath), "tmp-"+utils.GenerateToken()+"."+path.Base(outPath))

	if err := encode(tmpPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, outPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("finalize %s: %w", path.Base(outPath), err)
	}

	return nil
}

// persist (phase 4) writes everything process() computed to the database in
// a single transaction.
func (w *worker) persist(ctx context.Context, t *task) error {
	info := t.info
	plan := t.plan
	result := t.result
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

	if err := w.db.Transaction(func(tx *gorm.DB) error {
		if err := models.UpsertMedia(tx, media); err != nil {
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
			if err := upsertMediaURL(tx, media, purpose, file); err != nil {
				return fmt.Errorf("save media url (%s): %w", purpose, err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if info.isNewMedia {
		finalCachePath := utils.MediaCacheLeafPath(media.AlbumID, media.ID)
		if err := os.Rename(result.pendingCachePath, finalCachePath); err != nil {
			return fmt.Errorf("finalize cache directory: %w", err)
		}
	}

	if plan.needFaces && w.faces != nil {
		if err := w.faces.DetectFaces(w.db, media); err != nil {
			scanner_utils.ScannerError(ctx, "detect faces (%s): %s", t.path, err)
		}
	}

	return nil
}

func upsertMediaURL(tx *gorm.DB, media *models.Media, purpose models.MediaPurpose, file *encodedFile) error {
	url := &models.MediaURL{
		MediaID:     media.ID,
		MediaName:   file.name,
		Width:       file.width,
		Height:      file.height,
		Purpose:     purpose,
		ContentType: file.contentType,
		FileSize:    file.fileSize,
	}
	return models.UpsertMediaURL(tx, url)
}
