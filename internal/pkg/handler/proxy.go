package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type proxy struct {
	url *url.URL
}

// Proxy creates handler
func Proxy(url *url.URL) http.Handler {
	res := &proxy{}
	res.url = url
	return res
}

func (h *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctxSp, span := utils.StartSpan(r.Context(), "proxy.ServeHTTP", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	r = r.WithContext(ctxSp)

	rn, ctx := customContext(r)
	rn.URL.Host = h.url.Host
	rn.URL.Scheme = h.url.Scheme
	rn.Header.Set("X-Forwarded-Host", rn.Header.Get("Host"))
	otel.GetTextMapPropagator().Inject(r.Context(), propagation.HeaderCarrier(rn.Header))
	rn.Host = h.url.Host

	proxy := httputil.NewSingleHostReverseProxy(h.url)
	proxy.ModifyResponse = func(resp *http.Response) (err error) {
		ctx.ResponseCode = resp.StatusCode
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		return nil
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Error().Msgf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		ctx.ResponseCode = http.StatusBadGateway
		span.SetStatus(codes.Error, "Proxy error")
		span.RecordError(err)
	}
	proxy.ServeHTTP(w, rn)
}

func (h *proxy) Info(pr string) string {
	return pr + fmt.Sprintf("Proxy (%s)\n", h.url.String())
}
