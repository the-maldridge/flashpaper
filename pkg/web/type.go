package web

import (
	"time"
)

// Storage provides an interface to the storage system.  This can be
// provided by anything assuming it provides a mechanism for
// expiration.
type Storage interface {
	PutEx(string, interface{}, time.Duration) error
	Get(string) (interface{}, error)
}
