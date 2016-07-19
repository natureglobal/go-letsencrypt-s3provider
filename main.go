package main

import (
	"github.com/rsc/letsencrypt"
	"log"
)

func main() {
	var m letsencrypt.Manager
	cert, err := m.Cert("api.nature.global")
	if err != nil {
		log.Fatalf("Cert failed, error: %s", err)
	}
	log.Printf("Success, cert: %#v", cert)
}
