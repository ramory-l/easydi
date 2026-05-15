// Package lifecycle defines the optional Start/Close interfaces that easydi's
// generated Container honors over di:expose nodes.
//
// A node included in Container.Exposed() may implement Starter and/or Closer.
// Container.Start calls Start in topological (dependency-first) order;
// Container.Close calls Close in the reverse order. Neither interface is
// required: nodes that implement neither are simply skipped.
package lifecycle

import "context"

// Starter is implemented by exposed nodes that must be started after the
// dependency graph is built (background workers, queue consumers, schedulers).
type Starter interface {
	Start(ctx context.Context) error
}

// Closer is implemented by exposed nodes that must release resources on
// shutdown (flush buffers, close pools/clients).
type Closer interface {
	Close(ctx context.Context) error
}
