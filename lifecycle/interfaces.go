package lifecycle

// Component blocks until it stops serving. server.Server satisfies it.
type Component interface {
	Start() error
}
