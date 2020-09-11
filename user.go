package letsencrypts3provider

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/go-acme/lego/v4/registration"
)

// You'll need a user or account type that implements acme.User
type user struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

var _ registration.User = user{}

func newUser(email string) (user, error) {
	// Create a user. New accounts need an email and private key to start.
	const rsaKeySize = 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return user{}, err
	}
	u := user{
		Email: email,
		key:   privateKey,
	}
	return u, err
}

func (u user) GetEmail() string {
	return u.Email
}
func (u user) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u user) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
