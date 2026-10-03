package resolvers

// Helper functions for scanner.go, kept in a separate file (not the
// gqlgen-managed resolver file) since gqlgen's codegen doesn't recognize
// non-resolver top-level declarations and will otherwise comment them out
// on the next `go generate` as "unknown code".

import "github.com/photoview/photoview/api/scanner/scanner_queue"

// addAlbumToQueue is a seam over scanner_queue.AddAlbumToQueue so tests can
// exercise ScanAlbum's permission checks without touching the real,
// process-global scanner queue, which starts a background worker goroutine
// on first use.
var addAlbumToQueue = scanner_queue.AddAlbumToQueue
