package main

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-acme/lego/challenge"
)

var (
	s3cli      *s3.S3
	bucketName string
)

type s3UploadingProvider struct {
}

func init() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Fatalf("session.NewSession failed, error: %s", err)
	}
	bucketName = os.Getenv("AWS_LETSENCRYPT_S3PROVIDER_BUCKET")
	if bucketName == "" {
		log.Fatalf("AWS_LETSENCRYPT_S3PROVIDER_BUCKET required")
	}
	s3cli = s3.New(sess)
}

func NewS3UploadingProvider() challenge.Provider {
	return s3UploadingProvider{}
}

func (p s3UploadingProvider) Present(domain, token, keyAuth string) error {
	log.Printf("Present domain: %s\ntoken: %s\nkeyAuth: %s", domain, token, keyAuth)

	if _, err := s3cli.PutObject(&s3.PutObjectInput{
		ACL:         aws.String(s3.BucketCannedACLPrivate),
		Body:        strings.NewReader(keyAuth),
		ContentType: aws.String("text/plain"),
		Bucket:      aws.String(bucketName),
		Key:         aws.String(token),
	}); err != nil {
		log.Printf("Put: %s failed, error: %s", token, err)
		return err
	}
	return nil
}

func (p s3UploadingProvider) CleanUp(domain, token, keyAuth string) error {
	log.Printf("CleanUp domain: %s\ntoken: %s\nkeyAuth: %s", domain, token, keyAuth)

	if _, err := s3cli.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(token),
	}); err != nil {
		log.Printf("Del: %s failed, error: %s", token, err)
		return err
	}
	return nil
}
