package middleware

import (
	"bytes"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

//RequestLoggerFactory creates an HTTP request logging middleware for go-chi web framework.
//goland:noinspection GoUnusedExportedFunction
func RequestLoggerFactory(logResponseBody bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var rw = w
			var buf bytes.Buffer
			if logResponseBody {
				// Wrap the response to enable access to the final status code & response contents
				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

				// Tee the response to an additional buffer so we can look inside & log it
				ww.Tee(&buf)
				rw = ww
			}

			// Invoke target handler, but time the request
			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start)

			// Prepare log context
			logger := logrus.
				WithField("rid", GetRequestID(r.Context())).
				WithField("remoteAddr", r.RemoteAddr).
				WithField("proto", r.Proto).
				WithField("method", r.Method).
				WithField("uri", r.RequestURI).
				WithField("host", r.Host).
				WithField("requestHeaders", r.Header).
				WithField("responseHeaders", w.Header()).
				WithField("elapsed", duration)

			// Add HTTP response bytes, if requested to
			if logResponseBody {
				ww := rw.(middleware.WrapResponseWriter)
				logger = logger.
					WithField("status", ww.Status()).
					WithField("bytesWritten", ww.BytesWritten()).
					WithField("bytesOut", buf.String())
			}

			// Log it
			logger.Info("HTTP request completed")
		})
	}
}
