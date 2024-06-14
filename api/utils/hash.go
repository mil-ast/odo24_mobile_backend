package utils

import (
	"crypto/rand"
	"crypto/sha256"
)

const saltSize = 16

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	return salt, err
}

func GetPasswordHash(password []byte, salt []byte) ([]byte, error) {
	hasherNewPassword := sha256.New()
	_, err := hasherNewPassword.Write(salt)
	if err != nil {
		return nil, err
	}
	_, err = hasherNewPassword.Write(password)
	if err != nil {
		return nil, err
	}
	newHash := hasherNewPassword.Sum(nil)
	return newHash, nil
}
