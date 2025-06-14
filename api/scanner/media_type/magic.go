package media_type

import (
	"strings"
	"sync"

	magic "github.com/hosom/gomagic"
	"github.com/photoview/photoview/api/log"
)

var libmagic struct {
	mu       sync.Mutex
	libmagic *libMagic
	err      error
}

func init() {
	libmagic.libmagic, libmagic.err = newLibMagic()
	if libmagic.err != nil {
		libmagic.libmagic = nil
		log.Error(nil, "Init libmagic error.", "error", libmagic.err)
	}
}

// GetMediaType returns a media type of file `f`.
// This function is thread-safe.
func GetMediaType(f string) MediaType {
	libmagic.mu.Lock()
	defer libmagic.mu.Unlock()

	if libmagic.err != nil {
		log.Warn(nil, "GetMediaType() error.", "error", libmagic.err, "file", f)
		return TypeUnknown
	}

	mime, err := libmagic.libmagic.Type(f)
	if err != nil {
		log.Warn(nil, "GetMediaType() error.", "error", err, "file", f)
		return TypeUnknown
	}

	mime = strings.SplitN(mime, ";", 2)[0]

	return mediaType(mime)
}

// libMagic parses the magic code in a file and returns the media type.
type libMagic struct {
	magic *magic.Magic
}

// newLibMagic creates an instance of Magic.
func newLibMagic() (*libMagic, error) {
	// MAGIC_SYMLINK: Follow symlinks
	// MAGIC_MIME: Return MIME type string
	// MAGIC_ERROR: Handle errors in magic database
	// MAGIC_NO_CHECK_COMPRESS: Don't check for compressed files
	// MAGIC_NO_CHECK_ENCODING: Don't check for text encodings
	m, err := magic.Open(magic.MAGIC_SYMLINK | magic.MAGIC_MIME | magic.MAGIC_ERROR | magic.MAGIC_NO_CHECK_COMPRESS | magic.MAGIC_NO_CHECK_ENCODING)
	if err != nil {
		return nil, err
	}

	return &libMagic{
		magic: m,
	}, nil
}

// Close closes the library.
func (m *libMagic) Close() error {
	return m.magic.Close()
}

// Type returns the media type of the given file `f`. It returns TypeUnknown if the file is not a image nor a video.
// This function is not thread-safe.
func (m *libMagic) Type(f string) (string, error) {
	return m.magic.File(f)
}
