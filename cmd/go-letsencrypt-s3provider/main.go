package main

import (
	"log"
	"os"

	letsencrypts3provider "github.com/natureglobal/go-letsencrypt-s3provider"
)

func main() {
	if err := letsencrypts3provider.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.Fatal(err)
	}
}
