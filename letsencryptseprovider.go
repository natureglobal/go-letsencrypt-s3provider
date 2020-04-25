package letsencrypts3provider

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
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
	return Obtain(domains, email, directory, bucket, privkeyFile, certFile)
}

// Obtain server key and certificates
func Obtain(domains []string, email, directory, bucket, privkeyFile, certFile string) error {
	privatekeyF, err := os.Create(privkeyFile)
	if err != nil {
		return fmt.Errorf("Failed to Create: %s, error: %s", privkeyFile, err)
	}
	certF, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("Failed to Create: %s, error: %s", certFile, err)
	}

	user, err := NewUser(email)
	if err != nil {
		return fmt.Errorf("Failed to NewUser, error: %s", err)
	}

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(&lego.Config{
		HTTPClient: http.DefaultClient,
		CADirURL:   directory,
		User:       user,
		Certificate: lego.CertificateConfig{
			KeyType: certcrypto.RSA2048,
		},
	})
	if err != nil {
		return fmt.Errorf("Failed to NewClient, error: %s", err)
	}

	// New users will need to register
	if _, err := client.Registration.Register(registration.RegisterOptions{
		// The client has a URL to the current Let's Encrypt Subscriber
		// Agreement. The user will need to agree to it.
		TermsOfServiceAgreed: true,
	}); err != nil {
		return fmt.Errorf("Failed to Register, error: %s", err)
	}

	provider, err := NewS3UploadingProvider(bucket)
	if err != nil {
		return fmt.Errorf("Failed to NewS3UploadingPrivider: %w", err)
	}
	// We only use HTTP01
	if err := client.Challenge.SetHTTP01Provider(provider); err != nil {
		return fmt.Errorf("Failed to SetChallengeProvider failed, error: %s", err)
	}

	certificates, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
		// The acme library takes care of completing the challenges to obtain the certificate(s).
		// The domains must resolve to this machine or you have to use the DNS challenge.
		Bundle: false,
		// ELB doesn't support OCSP stapling
		MustStaple: false,
	})
	if err != nil {
		return fmt.Errorf("Failed to ObtainCertificate failed, failures: %s", err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	if _, err := privatekeyF.Write(certificates.PrivateKey); err != nil {
		return fmt.Errorf("Failed to Write: %s, error: %s", privkeyFile, err)
	}
	if _, err := certF.Write(certificates.Certificate); err != nil {
		return fmt.Errorf("Failed to Write: %s, error: %s", privkeyFile, err)
	}
	return nil
}
