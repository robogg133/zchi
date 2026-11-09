package zchi

import (
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// Logger is a middleware that logs incoming requests with zerolog including panics.
func Logger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				if r := recover(); r != nil && r != http.ErrAbortHandler {
					logger.Error().Interface("recover", r).Bytes("stack", debug.Stack()).Msg("incoming_request_panic")
					ww.WriteHeader(http.StatusInternalServerError)
				}
				l := logger.Info().
					Str("remote_addr", r.RemoteAddr).
					Str("path", r.URL.Path).
					Str("proto", r.Proto).
					Str("method", r.Method).
					Str("user_agent", r.UserAgent()).
					Int("status_code", ww.Status()).
					Int64("bytes_in", r.ContentLength).
					Int("bytes_out", ww.BytesWritten())
				if id := middleware.GetReqID(r.Context()); id != "" {
					l.Str("request_id", id)
				}

				l.Msg("incoming_request")
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
