package web

import (
	"context"
	"time"
)

// Storage provides an interface to the storage system.  This can be
// provided by anything assuming it provides a mechanism for
// expiration.
type Storage interface {
	Ping(context.Context) error

	PutEx(context.Context, string, interface{}, time.Duration) error
	Get(context.Context, string) (string, error)
	Del(context.Context, string) error
}
