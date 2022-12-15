package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/go-http-utils/headers"
)

const gzipEnc = "gzip"

var gzw *gzip.Writer

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipCompress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		contentEncoding := r.Header.Get(headers.ContentEncoding)
		if contentEncoding != "" && contentEncoding != gzipEnc {
			http.Error(w, "wrong encoding", http.StatusBadRequest)
			return
		}

		if contentEncoding == gzipEnc {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer gzr.Close()

			r.Body = io.NopCloser(gzr)
		}

		acceptEncoding := r.Header.Get(headers.AcceptEncoding)
		if !strings.Contains(acceptEncoding, gzipEnc) {
			next.ServeHTTP(w, r)
			return
		}

		if gzw == nil {
			gzw, err = gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			gzw.Reset(w)
		}

		w.Header().Set(headers.ContentEncoding, gzipEnc)
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzw}, r)
	}
}