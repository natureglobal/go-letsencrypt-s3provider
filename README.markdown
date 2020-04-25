A Challenge Provider for Let's Encrypt which Uploads the Token and KeyAuth to Amazon S3
=======================================================================================

## Why?

We're using a name server that cannot manipulate it's records over an open API, and we cannot use webroot on our production servers.

## Description

This is a custom Let's Encrypt challenge provider based on the webroot provider but behaves differently when challenged.

1. When requested to meet the challenge, this script creates a file in a S3 bucket (which is provided as `AWS_LETSENCRYPT_S3PROVIDER_BUCKET` environment variable) with `token` as the name and `keyAuth` as the content.
2. And removes the file on cleanup.
3. When Let's Encrypt fetches the token (ex: `example.com/.well-known/acme-challenge/xxxxxx`), another web application handles the request (use the `func Handler(w http.ResponseWriter, req *http.Request)` in `handler/`), fetches the file from the same S3 bucket and responds to Let's Encrypt's request.

## Usage

``` bash
# creates privatekey.pem and cert.pem
AWS_SECRET_KEY={SECRET_KEY} \
AWS_ACCESS_KEY_ID={ACCESS_KEY} \
AWS_LETSENCRYPT_S3PROVIDER_BUCKET={bucket name} \
go-letsencrypt-s3provider {email} {domain1,domain2,..} production privatekey.pem cert.pem
```
It is easy to use go-s3proxy to access the challenge files placed in s3.

```go
import "github.com/natureglobal/go-s3proxy"

var handler http.Handler = s3proxy.Proxy(os.Getenv("AWS_LETSENCRYPT_S3PROVIDER_BUCKET")

func Handle(w http.ResponseWriter, req *http.Request) {
	log.Printf("Func path: %s", req.URL.Path)
	handler.ServeHTTP(w, req)
}
```
