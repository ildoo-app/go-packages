package server

// Server is an HTTP listener that serves a Connect (or other) handler.
// The value returned by New satisfies it.
type Server interface {
	Start() error
}
