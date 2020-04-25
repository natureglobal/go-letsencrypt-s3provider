package letsencrypts3provider

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-acme/lego/challenge"
)

type s3UploadingProvider struct {
	bucket string
	svc    *s3.S3
}

var _ challenge.Provider = s3UploadingProvider{}

const defaultRegion = "us-east-1"

func newS3() (*s3.S3, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	if *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String(defaultRegion)})
	}
	return s3.New(sess), nil
}

func newS3UploadingProvider(bucket string) (challenge.Provider, error) {
	svc, err := newS3()
	if err != nil {
		return nil, err
	}
	return s3UploadingProvider{
		svc:    svc,
		bucket: bucket,
	}, nil
}

func (p s3UploadingProvider) Present(domain, token, keyAuth string) error {
	log.Printf("Present domain: %s\ntoken: %s\nkeyAuth: %s", domain, token, keyAuth)

	if _, err := p.svc.PutObject(&s3.PutObjectInput{
		ACL:         aws.String(s3.BucketCannedACLPrivate),
		Body:        strings.NewReader(keyAuth),
		ContentType: aws.String("text/plain"),
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(token),
	}); err != nil {
		log.Printf("Put: %s failed, error: %s", token, err)
		return err
	}
	return nil
}

func (p s3UploadingProvider) CleanUp(domain, token, keyAuth string) error {
	log.Printf("CleanUp domain: %s\ntoken: %s\nkeyAuth: %s", domain, token, keyAuth)

	if _, err := p.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(token),
	}); err != nil {
		log.Printf("Del: %s failed, error: %s", token, err)
		return err
	}
	return nil
}
