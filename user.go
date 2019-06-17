package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/go-acme/lego/registration"
)

// You'll need a user or account type that implements acme.User
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func NewUser(email string) (User, error) {
	// Create a user. New accounts need an email and private key to start.
	const rsaKeySize = 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return User{}, err
	}
	user := User{
		Email: email,
		key:   privateKey,
	}
	return user, err
}

func (u User) GetEmail() string {
	return u.Email
}
func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
