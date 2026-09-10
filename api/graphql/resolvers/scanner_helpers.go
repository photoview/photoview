package resolvers

// Helper functions for scanner.go, kept in a separate file (not the
// gqlgen-managed resolver file) since gqlgen's codegen doesn't recognize
// non-resolver top-level declarations and will otherwise comment them out
// on the next `go generate` as "unknown code".

import "github.com/photoview/photoview/api/scanner/scanner_queue"

// getScannerQueueStatus is a seam over scanner_queue.GetQueueStatus so
// tests can exercise ScannerQueueStatus's permission filtering without
// touching the real, process-global scanner queue.
var getScannerQueueStatus = scanner_queue.GetQueueStatus

// cancelScanJob is a seam over scanner_queue.CancelJob so tests can
// exercise CancelScanJob's permission checks without touching the real,
// process-global scanner queue.
var cancelScanJob = scanner_queue.CancelJob

// cancelAllScanJobs is a seam over scanner_queue.CancelAllJobs so tests can
// exercise CancelAllScanJobs's admin path without touching the real,
// process-global scanner queue.
var cancelAllScanJobs = scanner_queue.CancelAllJobs
