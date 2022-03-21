package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values or stored/retrieved.
const KeyValues ctxKey = 1

// Values carries information about each request.
type Values struct {
	StatusCode int
	Start      time.Time
	TraceID    string
}

// Handler is the signature that all application handlers will implement.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// Router is the entry point for all web applications.
type Router struct {
	mux      *chi.Mux
	Log      *log.Logger
	mw       []Middleware
	och      *ochttp.Handler
	shutdown chan os.Signal
}

// NewRouter knows how to construct internal state for a Router.
func NewRouter(shutdown chan os.Signal, logger *log.Logger, mw ...Middleware) *Router {

	router := Router{
		mux:      chi.NewRouter(),
		Log:      logger,
		mw:       mw,
		shutdown: shutdown,
	}

	// Create an OpenCensus HTTP handler which wraps the router. This will start
	// the initial span and annotate it with information about the request/response.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/
	router.och = &ochttp.Handler{
		Handler:     router.mux,
		Propagation: &tracecontext.HTTPFormat{},
	}

	return &router
}

// Handler connects a method and URL pattern to a particular application handler.
func (router *Router) Handler(method, pattern string, h Handler, mw ...Middleware) {
	// First wrap handler specific middleware around this handler.
	h = wrapMiddleware(mw, h)

	// Add the application's general middleware to the handler chain
	h = wrapMiddleware(router.mw, h)

	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(r.Context(), "internal.plaform.web")
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID.String(),
			Start:   time.Now(),
		}

		ctx = context.WithValue(ctx, KeyValues, &v)

		if err := h(ctx, w, r); err != nil {
			router.Log.Printf("%s : Unhandled error %+v", v.TraceID, err)
			if IsShutdown(err) {
				router.SignalShutdown()
			}
		}
	}
	router.mux.MethodFunc(method, pattern, fn)
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	router.och.ServeHTTP(writer, request)
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (router *Router) SignalShutdown() {
	router.Log.Println("error returned from handler indicated integrity issue, shutting down service")
	router.shutdown <- syscall.SIGSTOP
}
