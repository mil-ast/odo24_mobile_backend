package services

import "errors"

var (
	ErrorUnauthorize  = errors.New("unauthorize")
	ErrorNoPermission = errors.New("no_permission")
)
