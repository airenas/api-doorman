package service

import (
	"context"
	slog "log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/pkg/errors"
)

type (
	//IPManager manages IP in DB
	IPManager interface {
		CheckCreateIPKey(ctx context.Context, ip string, limit float64) (string /*key ID*/, error)
	}

	//HandlerWrap for check if handler valid
	HandlerWrap interface {
		Valid(r *http.Request) bool
		Handler() http.Handler
		Info() string
		Name() string
		Priority() int
	}

	//Data is service operation data
	Data struct {
		Port     int
		Handlers []HandlerWrap
	}
)

type mainHandler struct {
	data *Data
}

// StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	log.Info().Msgf("Starting HTTP service at %d", data.Port)
	h, err := newMainHandler(data)
	if err != nil {
		return errors.Wrap(err, "can't init handlers")
	}

	portStr := strconv.Itoa(data.Port)

	logHandlers(getInfo(data.Handlers))

	gracehttp.SetLogger(slog.New(goapp.Log, "", 0))

	return gracehttp.Serve(&http.Server{
		Addr:        ":" + portStr,
		IdleTimeout: 10 * time.Minute, ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout: 8 * time.Minute, WriteTimeout: 15 * time.Minute,
		Handler: h,
	})
}

func logHandlers(info string) {
	for _, s := range strings.Split(info, "\n") {
		log.Info().Msg(s)
	}
}

func newMainHandler(data *Data) (http.Handler, error) {
	res := &mainHandler{}
	if len(data.Handlers) == 0 {
		return nil, errors.New("No handlers")
	}
	res.data = data
	sort.Slice(res.data.Handlers, func(i, j int) bool { return res.data.Handlers[i].Priority() > res.data.Handlers[j].Priority() })
	return res, nil
}

func (h *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := initSpan(ctx)
	defer span.End()

	for _, hi := range h.data.Handlers {
		if hi.Valid(r) {
			span.SetAttributes(attribute.String("handler.name", hi.Name()))
			log.Ctx(ctx).Info().Msg("Handling with " + hi.Name())
			hi.Handler().ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}
	log.Ctx(ctx).Error().Str("path", r.URL.Path).Msg("no handler")
	//serve not found
	http.NotFound(w, r)
}

func initSpan(ctx context.Context) (context.Context, trace.Span) {
	ctx, span := otel.Tracer("api-doorman").Start(ctx, "main handler", trace.WithSpanKind(trace.SpanKindServer))
	ctx = loggerWithTrace(ctx).WithContext(ctx)
	return ctx, span
}

func loggerWithTrace(ctx context.Context) zerolog.Logger {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	if traceID == "" {
		return log.Logger
	}
	return log.With().Str("traceID", traceID).Logger()
}

func getInfo(handlers []HandlerWrap) string {
	sb := strings.Builder{}
	for _, h := range handlers {
		sb.WriteString(h.Info())
		sb.WriteString("\n")
	}
	return sb.String()
}
