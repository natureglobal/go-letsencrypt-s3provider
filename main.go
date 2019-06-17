package main

import (
	"log"
	"net/http"
	"os"
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

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	if len(os.Args) != 6 {
		log.Fatalf("Usage: %s {email} {domain1,domain2,..} {production|staging} {out privatekey} {out cert}", os.Args[0])
	}
	email := os.Args[1]
	domains := strings.Split(os.Args[2], ",")
	directory := os.Args[3]
	if directory == "production" {
		directory = directoryURL
	} else {
		directory = stagingDirectoryURL
	}
	privatekeyFilename := os.Args[4]
	privatekeyFile, err := os.Create(privatekeyFilename)
	if err != nil {
		log.Fatalf("Failed to Create: %s, error: %s", privatekeyFilename, err)
	}
	certFilename := os.Args[5]
	certFile, err := os.Create(certFilename)
	if err != nil {
		log.Fatalf("Failed to Create: %s, error: %s", certFilename, err)
	}

	log.Printf("Using directory: %s", directory)

	user, err := NewUser(email)
	if err != nil {
		log.Fatalf("Failed to NewUser, error: %s", err)
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
		log.Fatalf("Failed to NewClient, error: %s", err)
	}

	// New users will need to register
	if _, err := client.Registration.Register(registration.RegisterOptions{
		// The client has a URL to the current Let's Encrypt Subscriber
		// Agreement. The user will need to agree to it.
		TermsOfServiceAgreed: true,
	}); err != nil {
		log.Fatalf("Failed to Register, error: %s", err)
	}

	provider := NewS3UploadingProvider()
	// We only use HTTP01
	if err := client.Challenge.SetHTTP01Provider(provider); err != nil {
		log.Fatalf("Failed to SetChallengeProvider failed, error: %s", err)
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
		log.Fatalf("Failed to ObtainCertificate failed, failures: %s", err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	// log.Printf("certificates: %#v\n", certificates)
	if _, err := privatekeyFile.Write(certificates.PrivateKey); err != nil {
		log.Fatalf("Failed to Write: %s, error: %s", privatekeyFilename, err)
	}
	if _, err := certFile.Write(certificates.Certificate); err != nil {
		log.Fatalf("Failed to Write: %s, error: %s", privatekeyFilename, err)
	}
}
