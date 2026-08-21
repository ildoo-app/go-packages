package server

import (
	"errors"
	"strings"
)

// ServicePath is the URL prefix a Connect handler is mounted at,
// e.g. "/ildoo.merchants.v1.MerchantService/".
type ServicePath string

func (p ServicePath) validate() error {
	if p == "" {
		return errors.New("server: service path is empty")
	}
	if !strings.HasPrefix(p.String(), "/") {
		return errors.New("server: service path must start with /")
	}
	return nil
}

func (p ServicePath) String() string {
	return string(p)
}
