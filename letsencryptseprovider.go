package letsencrypts3provider

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

const (
	stagingDirectoryURL = lego.LEDirectoryStaging
	directoryURL        = lego.LEDirectoryProduction
)

var helpReg = regexp.MustCompile(`^--?h(?:elp)?$`)

func help() string {
	return "Usage: go-letsencrypt-s3provider {email} {domain1,domain2,..} {production|staging} {out privatekey} {out cert}"
}

// Run the letsencrypts3provider cli
func Run(argv []string, stdout, stderr io.Writer) error {
	if len(argv) != 5 || helpReg.MatchString(argv[0]) {
		fmt.Fprintln(stdout, help())
	}
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	email := argv[0]
	domains := strings.Split(argv[1], ",")
	directory := argv[2]
	privkeyFile := argv[3]
	certFile := argv[4]

	if directory == "production" {
		directory = directoryURL
	} else {
		directory = stagingDirectoryURL
	}

	bucket := os.Getenv("AWS_LETSENCRYPT_S3PROVIDER_BUCKET")
	if bucket == "" {
		return fmt.Errorf("AWS_LETSENCRYPT_S3PROVIDER_BUCKET required")
	}
	certificates, err := Obtain(&ObtainRequest{
		Domains:   domains,
		Directory: directory,
		Email:     email,
		Bucket:    bucket,
	})
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(privkeyFile, certificates.PrivateKey, 0644); err != nil {
		return err
	}
	return ioutil.WriteFile(certFile, certificates.Certificate, 0644)
}

type ObtainRequest struct {
	Domains        []string
	Directory      string
	Email          string
	Bucket         string
	PreferredChain string
}

// Obtain server key and certificates
func Obtain(ob *ObtainRequest) (*certificate.Resource, error) {
	u, err := newUser(ob.Email)
	if err != nil {
		return nil, fmt.Errorf("Failed to NewUser, error: %s", err)
	}
	directory := ob.Directory
	if directory == "" {
		directory = directoryURL
	}
	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(&lego.Config{
		HTTPClient: http.DefaultClient,
		CADirURL:   directory,
		User:       u,
		Certificate: lego.CertificateConfig{
			KeyType: certcrypto.RSA2048,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to NewClient, error: %s", err)
	}

	// New users will need to register
	if _, err := client.Registration.Register(registration.RegisterOptions{
		// The client has a URL to the current Let's Encrypt Subscriber
		// Agreement. The user will need to agree to it.
		TermsOfServiceAgreed: true,
	}); err != nil {
		return nil, fmt.Errorf("Failed to Register, error: %s", err)
	}

	provider, err := newS3UploadingProvider(ob.Bucket)
	if err != nil {
		return nil, fmt.Errorf("Failed to NewS3UploadingPrivider: %w", err)
	}
	// We only use HTTP01
	if err := client.Challenge.SetHTTP01Provider(provider); err != nil {
		return nil, fmt.Errorf("Failed to SetChallengeProvider failed, error: %s", err)
	}

	return client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: ob.Domains,
		// The acme library takes care of completing the challenges to obtain the certificate(s).
		// The domains must resolve to this machine or you have to use the DNS challenge.
		Bundle: false,
		// ELB doesn't support OCSP stapling
		MustStaple:     false,
		PreferredChain: ob.PreferredChain,
	})
}
