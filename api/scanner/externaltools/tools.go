// Package externaltools provides wrappers for tools outside the Go runtime.
// These tools are provided by the runtime environment. Some require initialization
// and must be cleaned up properly when finished.
// Packages under externaltools should have as few dependencies as possible to avoid cycles.
package externaltools
