package services

import "errors"

var (
	ErrorUnauthorize          = errors.New("unauthorize")
	ErrorNoPermission         = errors.New("no_permission")
	ErrorParsingRSAPublicKey  = errors.New("error parsing RSA public key")
	ErrorParsingRSAPrivateKey = errors.New("error parsing RSA private key")
	ErrorSigningJwtToken      = errors.New("error signing token")
)
