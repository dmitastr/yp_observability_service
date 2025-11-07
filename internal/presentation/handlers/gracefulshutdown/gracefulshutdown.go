package gracefulshutdown

import (
	"context"
	"net/http"
	"os"
)

// Handler wraps each request with context that listens to shutdown signal
type Handler struct {
	cancelCh chan os.Signal
}

// NewGracefulShutdownHandler creates new [Handler] with parent context
func NewGracefulShutdownHandler(cancelCh chan os.Signal) *Handler {
	return &Handler{cancelCh: cancelCh}
}

// Handle wrap request with context that listens to shutdown signal
func (g *Handler) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := g.WithCancellation(r.Context(), g.cancelCh)
		defer cancel()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// WithCancellation creates new context that will listten to shutdown signal and invoke cancel
func (g *Handler) WithCancellation(ctx context.Context, ch chan os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
			return
		}
	}()
	return ctx, cancel
}
