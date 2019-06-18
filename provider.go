package main

import (
	"log"
	"os"

	"github.com/go-acme/lego/challenge"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

var (
	Bucket *s3.Bucket
)

type s3UploadingProvider struct {
}

func init() {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatalf("EnvAuth failed, error: %s", err)
	}
	bucket := os.Getenv("AWS_LETSENCRYPT_S3PROVIDER_BUCKET")
	if bucket == "" {
		log.Fatalf("AWS_LETSENCRYPT_S3PROVIDER_BUCKET required")
	}
	client := s3.New(auth, aws.USEast)
	Bucket = client.Bucket(bucket)
}

func NewS3UploadingProvider() challenge.Provider {
	return s3UploadingProvider{}
}

func (p s3UploadingProvider) Present(domain, token, keyAuth string) error {
	log.Printf("Present domain: %s\ntoken: %s\nkeyAuth: %s", domain, token, keyAuth)

	if err := Bucket.Put(token, []byte(keyAuth), "text/plain", s3.Private); err != nil {
		log.Printf("Put: %s failed, error: %s", token, err)
	}

	return nil
}

func (p s3UploadingProvider) CleanUp(domain, token, keyAuth string) error {
	log.Printf("CleanUp domain: %s\ntoken: %s\nkeyAuth: %s", domain, token, keyAuth)

	if err := Bucket.Del(token); err != nil {
		log.Printf("Del: %s failed, error: %s", token, err)
	}
	return nil
}
