package handler

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mash/go-s3proxy"
)

const (
	ExpiresInterval time.Duration = 1 * time.Minute
	UserAgent       string        = "s3file/1.0"
)

var handler http.Handler

func init() {
	bucket := os.Getenv("AWS_LETSENCRYPT_S3PROVIDER_BUCKET")
	if bucket == "" {
		panic("AWS_LETSENCRYPT_S3PROVIDER_BUCKET required")
	}
	handler = s3proxy.Proxy(bucket)
}

func Handler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Func path: %s", req.URL.Path)
	handler.ServeHTTP(w, req)
}
