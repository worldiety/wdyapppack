package run

import (
	"sync"
)

// Cleanups run when the process is interrupted, so that what an adapter set up
// for one command — a temporary keychain holding a signing certificate — does
// not outlive it. Deferred functions do not run when a signal ends the
// process; these do, as long as the program calls Interrupted from its signal
// handler.
var (
	cleanupMu sync.Mutex
	cleanups  = map[int]func(){}
	nextID    int
)

// OnInterrupt registers a function to run if the process is interrupted. The
// function returned unregisters it; call it once the cleanup has happened the
// ordinary way.
func OnInterrupt(fn func()) (done func()) {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	id := nextID
	nextID++
	cleanups[id] = fn
	return func() {
		cleanupMu.Lock()
		defer cleanupMu.Unlock()
		delete(cleanups, id)
	}
}

// Interrupted runs every registered cleanup once.
func Interrupted() {
	cleanupMu.Lock()
	fns := make([]func(), 0, len(cleanups))
	for id, fn := range cleanups {
		fns = append(fns, fn)
		delete(cleanups, id)
	}
	cleanupMu.Unlock()
	for _, fn := range fns {
		fn()
	}
}
