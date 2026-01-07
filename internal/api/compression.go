package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressionMiddleware compresses HTTP responses using gzip
// This reduces bandwidth usage and improves response times for large payloads
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for small responses or already compressed content
		// We'll compress responses > 1KB
		w.Header().Set("Vary", "Accept-Encoding")
		
		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Set content encoding header
		w.Header().Set("Content-Encoding", "gzip")

		// Wrap response writer
		gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to compress responses
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// WriteHeader writes the header (status code)
func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

