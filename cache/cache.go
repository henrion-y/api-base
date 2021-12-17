package cache

import (
	"errors"
	"time"
)

type Cache interface {
	Get(key string, resultPtr interface{}) error
	BatchGet(keys []string, resultsPtr interface{}) error
	Set(key string, value interface{}, expiry time.Duration) error
	SetNX(key string, value interface{}, expiry time.Duration) error
	Delete(key string) error
	BatchDelete(keys []string) error
}

var ErrNil = errors.New("nil returned")
